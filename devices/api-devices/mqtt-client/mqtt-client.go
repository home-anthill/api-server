package mqtt_client

import (
  "crypto/tls"
  "crypto/x509"
  "fmt"
  mqtt "github.com/eclipse/paho.mqtt.golang"
  "os"
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

//func PublishMessage(msg mqtt.Message) {
//  fmt.Printf("Topic: %s\n", msg.Topic())
//  fmt.Printf("Payload: %s\n", msg.Payload())
//  uuid := strings.Split(msg.Topic(), "/")[1]
//  // do a call to a function to publish `msg.Payload()`
//}

func NewTLSConfig() *tls.Config {
  // Import trusted certificates from CAfile.pem.
  // Alternatively, manually add CA certificates to
  // default openssl CA bundle.
  certpool := x509.NewCertPool()
  pemCerts, err := os.ReadFile(os.Getenv("MQTT_CA_FILE"))
  if err == nil {
    certpool.AppendCertsFromPEM(pemCerts)
  }

  // Import client certificate/key pair
  cert, err := tls.LoadX509KeyPair(os.Getenv("MQTT_CERT_FILE"), os.Getenv("MQTT_KEY_FILE"))
  if err != nil {
    panic(err)
  }

  // Just to print out the client certificate..
  cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
  if err != nil {
    panic(err)
  }
  fmt.Println(cert.Leaf)

  // Create tls.Config with desired tls properties
  return &tls.Config{
    // RootCAs = certs used to verify server cert.
    RootCAs: certpool,
    // ClientAuth = whether to request cert from server.
    // Since the server is set up for SSL, this happens
    // anyway.
    ClientAuth: tls.NoClientCert,
    // ClientCAs = certs used to validate client cert.
    ClientCAs: nil,
    // InsecureSkipVerify = verify that cert contents
    // match server. IP matches what is in cert etc.
    InsecureSkipVerify: true,
    // Certificates = list of certs client sends to server.
    Certificates: []tls.Certificate{cert},
  }
}

func InitMqtt() {
  //mqtt.DEBUG = log.New(os.Stdout, "", 0)
  //mqtt.ERROR = log.New(os.Stdout, "", 0)
  mqttUrl := os.Getenv("MQTT_URL") + ":" + os.Getenv("MQTT_PORT")

  opts := mqtt.NewClientOptions()
  opts.SetKeepAlive(5 * time.Second)
  opts.SetPingTimeout(2 * time.Second)
  opts.AddBroker(mqttUrl)
  opts.SetDefaultPublishHandler(defaultHandler)

  if os.Getenv("MQTT_TLS") == "true" {
    tlsConfig := NewTLSConfig()
    opts.SetClientID("apiDevices").SetTLSConfig(tlsConfig)
  } else {
    opts.SetClientID("apiDevices")
  }

  //c = mqtt.NewClient(opts)
  //if token := c.Connect(); token.Wait() && token.Error() != nil {
  //  panic(token.Error())
  //}
  //
  //// Subscribe to devices notification with new values
  //c.Subscribe("devices/+/notify/onoff", qos, func(client mqtt.Client, msg mqtt.Message) {
  //  fmt.Println("Received a onOff message via MQTT")
  //  //PublishMessage(msg)
  //})
  //c.Subscribe("devices/+/notify/temperature", qos, func(client mqtt.Client, msg mqtt.Message) {
  //  fmt.Println("Received a temperature message via MQTT")
  //  //PublishMessage(msg)
  //})
  //c.Subscribe("devices/+/notify/mode", qos, func(client mqtt.Client, msg mqtt.Message) {
  //  fmt.Println("Received a mode message via MQTT")
  //  //PublishMessage(msg)
  //})
  //c.Subscribe("devices/+/notify/fanMode", qos, func(client mqtt.Client, msg mqtt.Message) {
  //  fmt.Println("Received a fanMode message via MQTT")
  //  //PublishMessage(msg)
  //})
  //c.Subscribe("devices/+/notify/fanSpeed", qos, func(client mqtt.Client, msg mqtt.Message) {
  //  fmt.Println("Received a fanSpeed message via MQTT")
  //  //PublishMessage(msg)
  //})
  time.Sleep(6 * time.Second)
}
