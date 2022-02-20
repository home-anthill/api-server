package mqtt_client

import (
  amqpPublisher "api-devices/amqp-publisher"
  "fmt"
  mqtt "github.com/eclipse/paho.mqtt.golang"
  "strings"
  "time"
)

const qos byte = 0

var c mqtt.Client

var defaultHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
  fmt.Printf("---UNKNOWN TOPIC---")
  fmt.Printf("MessageID: %d\n", msg.MessageID())
  fmt.Printf("Topic: %s\n", msg.Topic())
  fmt.Printf("Payload: %s\n", msg.Payload())
  fmt.Printf("------------------")
}

func SendOnOff(uuid string, messageJSON []byte) mqtt.Token {
  fmt.Println("SendOnOff - publishing message...")
  return c.Publish("devices/"+uuid+"/onoff", qos, false, messageJSON)
}
func SendTemperature(uuid string, messageJSON []byte) mqtt.Token {
  fmt.Println("SendTemperature - publishing message...")
  return c.Publish("devices/"+uuid+"/temperature", qos, false, messageJSON)
}
func SendMode(uuid string, messageJSON []byte) mqtt.Token {
  fmt.Println("SendMode - publishing message...")
  return c.Publish("devices/"+uuid+"/mode", qos, false, messageJSON)
}
func SendFanMode(uuid string, messageJSON []byte) mqtt.Token {
  fmt.Println("SendFanMode - publishing message...")
  return c.Publish("devices/"+uuid+"/fanMode", qos, false, messageJSON)
}
func SendFanSpeed(uuid string, messageJSON []byte) mqtt.Token {
  fmt.Println("SendFanSpeed - publishing message...")
  return c.Publish("devices/"+uuid+"/fanSpeed", qos, false, messageJSON)
}

func PublishMessage(msg mqtt.Message) {
  fmt.Printf("Topic: %s\n", msg.Topic())
  fmt.Printf("Payload: %s\n", msg.Payload())
  uuid := strings.Split(msg.Topic(), "/")[1]
  amqpPublisher.Publish(uuid, msg.Payload())
}

func InitMqtt() {
  //mqtt.DEBUG = log.New(os.Stdout, "", 0)
  //mqtt.ERROR = log.New(os.Stdout, "", 0)
  opts := mqtt.NewClientOptions().AddBroker("tcp://eclipse-mosquitto:1883").SetClientID("apiDevices")
  opts.SetKeepAlive(2 * time.Second)
  opts.SetDefaultPublishHandler(defaultHandler)
  opts.SetPingTimeout(1 * time.Second)

  c = mqtt.NewClient(opts)
  if token := c.Connect(); token.Wait() && token.Error() != nil {
    panic(token.Error())
  }

  // Subscribe to devices notification with new values
  c.Subscribe("devices/+/notify/onoff", qos, func(client mqtt.Client, msg mqtt.Message) {
    fmt.Println("Received a onOff message via MQTT")
    PublishMessage(msg)
  })
  c.Subscribe("devices/+/notify/temperature", qos, func(client mqtt.Client, msg mqtt.Message) {
    fmt.Println("Received a temperature message via MQTT")
    PublishMessage(msg)
  })
  c.Subscribe("devices/+/notify/mode", qos, func(client mqtt.Client, msg mqtt.Message) {
    fmt.Println("Received a mode message via MQTT")
    PublishMessage(msg)
  })
  c.Subscribe("devices/+/notify/fanMode", qos, func(client mqtt.Client, msg mqtt.Message) {
    fmt.Println("Received a fanMode message via MQTT")
    PublishMessage(msg)
  })
  c.Subscribe("devices/+/notify/fanSpeed", qos, func(client mqtt.Client, msg mqtt.Message) {
    fmt.Println("Received a fanSpeed message via MQTT")
    PublishMessage(msg)
  })
  time.Sleep(6 * time.Second)
}
