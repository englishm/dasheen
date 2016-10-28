package main

import (
	"encoding/json"
	"flag"
	mqtt "./vendor/git.eclipse.org/r/paho/org.eclipse.paho.mqtt.golang"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type bathroomStatus struct {
	Upstairs   string
	Downstairs string
	Downstairs2 string
}

var status = bathroomStatus{
	Upstairs:   "unknown",
	Downstairs: "unknown",
	Downstairs2: "unknown",
}

var connections map[*websocket.Conn]bool

func onMessageReceived(client *mqtt.MqttClient, message mqtt.Message) {
	t := string(message.Topic())
	msg := string(message.Payload())
	log.Println(t + " " + msg)
	if t == "callaloo/upstairs" {
		status.Upstairs = msg
	}
	if t == "callaloo/downstairs" {
		status.Downstairs = msg
	}
	if t == "callaloo/downstairs2" {
		status.Downstairs2 = msg
	}
	jsonStatus, err := json.Marshal(status)
	if err != nil {
		log.Println("error:", err)
	}

	wsMessage := []byte(jsonStatus)
	sendAll(wsMessage)
}

func sendAll(msg []byte) {
	for conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			delete(connections, conn)
			conn.Close()
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	log.Println("Successfully upgraded connection")
	connections[conn] = true

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			delete(connections, conn)
			conn.Close()
			return
		}
		log.Println(string(msg))
		sendAll(msg)
	}
}

func mqttSetup(broker *string, clientid *string, topic *string, qos *int) {
	opts := mqtt.NewClientOptions().SetBroker(*broker).SetClientId(*clientid).SetCleanSession(true).SetTraceLevel(mqtt.Off)
	client := mqtt.NewClient(opts)
	_, err := client.Start()
	if err != nil {
		panic(err)
	} else {
		log.Printf("Connected as %s to %s\n", *clientid, *broker)
	}
	filter, e := mqtt.NewTopicFilter(*topic, byte(*qos))
	if e != nil {
		log.Fatal(e)
	}
	client.StartSubscription(onMessageReceived, filter)

}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	jsonStatus, err := json.Marshal(status)
	if err != nil {
		log.Println("error:", err)
	}
	w.Write(jsonStatus)
}

func webSetup(dir *string, iface *string, port *string) {
	connections = make(map[*websocket.Conn]bool)
	fs := http.Dir(*dir)
	fileHandler := http.FileServer(fs)
	http.Handle("/", fileHandler)
	http.HandleFunc("/status.json", jsonHandler)
	http.HandleFunc("/ws", wsHandler)
	err := http.ListenAndServe(*iface+":"+*port, nil)
	if err != nil {
		log.Println("error:", err)
	}
}

func main() {
	var (

		// MQTT configuration

		broker   = flag.String("broker", "tcp://10.138.123.50:1883", "mqtt broker")
		clientid = flag.String("clientid", "dasheen", "mqtt clientid")
		topic    = flag.String("topic", "callaloo/#", "mqtt topic name")
		qos      = flag.Int("qos", 1, "mqtt quality of service level")

		// Web and websockets configuration

		dir   = flag.String("directory", "web/", "directory of web files")
		port  = flag.String("port", "4102", "port to run on")
		iface = flag.String("iface", "0.0.0.0", "network interface to bind to")
	)
	flag.Parse()
	mqttSetup(broker, clientid, topic, qos)
	webSetup(dir, iface, port)
}
