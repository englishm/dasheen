package main

import (
    "fmt"
    "flag"
    "os"
    "log"
    "net/http"
    "encoding/json"
    "github.com/gorilla/websocket"
    mqtt "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

type BathroomStatus struct {
  Upstairs string
  Downstairs string
}

var status = BathroomStatus{
  Upstairs: "unknown",
  Downstairs: "unknown",
}

var connections map[*websocket.Conn]bool

func onMessageReceived(client *mqtt.MqttClient, message mqtt.Message) {
  fmt.Printf("Received message on topic: %s\n", message.Topic())
  t := string(message.Topic())
  msg := string(message.Payload())
  fmt.Printf("Message: %s\n", message.Payload())
  if t == "callaloo/upstairs" {
    status.Upstairs = msg
  }
  if t == "callaloo/downstairs" {
    status.Downstairs = msg
  }
  fmt.Printf("upstairs: %s\n", status.Upstairs)
  fmt.Printf("downstairs: %s\n", status.Downstairs)
  jsonStatus, err := json.Marshal(status)
  if err != nil {
    fmt.Println("error:", err)
  }
  os.Stdout.Write(jsonStatus)

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

func mqttSetup() {
  broker := flag.String("broker", "tcp://10.138.123.180:1883", "MQTT broker")
  clientid := flag.String("clientid", "dasheen", "Clientid")
  topic := flag.String("topic", "callaloo/#", "Topic name")
  qos := flag.Int("qos", 1, "QoS")
  flag.Parse()
  opts := mqtt.NewClientOptions().SetBroker(*broker).SetClientId(*clientid).SetCleanSession(true).SetTraceLevel(mqtt.Off)
  client := mqtt.NewClient(opts)
  _, err := client.Start()
  if err != nil {
    panic(err)
  } else {
    fmt.Printf("Connected as %s to %s\n", *clientid, *broker)
  }
  filter, e := mqtt.NewTopicFilter(*topic, byte(*qos))
  if e != nil {
    fmt.Println(e)
    os.Exit(1)
  }
  client.StartSubscription(onMessageReceived, filter)

}

func hello(w http.ResponseWriter, r *http.Request) { 
  fmt.Fprintf(w, "Hello.\n upstairs: " + status.Upstairs + "\n downstairs: " + status.Downstairs)
}

func main() {
  mqttSetup()
  dir := flag.String("directory", "web/", "directory of web files")
  flag.Parse()
  connections = make(map[*websocket.Conn]bool)
  fs := http.Dir(*dir)
  fileHandler := http.FileServer(fs)
  http.Handle("/", fileHandler)
  http.HandleFunc("/hello",hello)
  http.HandleFunc("/ws", wsHandler)
  http.ListenAndServe("0.0.0.0:80", nil)
  // web.Get("/(.*)", hello)
  // web.Run("0.0.0.0:9999")
}

