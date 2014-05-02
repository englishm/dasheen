package main

import (
    "github.com/hoisie/web"
    "fmt"
    "flag"
    "os"
    mqtt "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

var (
  upstairs = "unknown"
  downstairs = "unknown"
)

func onMessageReceived(client *mqtt.MqttClient, message mqtt.Message) {
  fmt.Printf("Received message on topic: %s\n", message.Topic())
  t := string(message.Topic())
  msg := string(message.Payload())
  fmt.Printf("Message: %s\n", message.Payload())
  if t == "callaloo/upstairs" {
    upstairs = msg
  }
  if t == "callaloo/downstairs" {
    downstairs = msg
  }
  fmt.Printf("upstairs: %s\n", upstairs)
  fmt.Printf("downstairs: %s\n", downstairs)
}

func setup() {
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

func hello(val string) string { 
  return "Hello, " + val + ".<br/>\n<br/>\n upstairs: " + upstairs + "<br/>\n downstairs: " + downstairs
}

func main() {
    setup()
    web.Get("/(.*)", hello)
    web.Run("0.0.0.0:9999")
}

