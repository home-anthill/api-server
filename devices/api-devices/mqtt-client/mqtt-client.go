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

var mqttClient mqtt.Client

var defaultHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
  fmt.Printf("---UNKNOWN TOPIC---")
  fmt.Printf("MessageID: %d\n", msg.MessageID())
  fmt.Printf("Topic: %s\n", msg.Topic())
  fmt.Printf("Payload: %s\n", msg.Payload())
  fmt.Printf("------------------")
}

func SendValues(uuid string, messageJSON []byte) mqtt.Token {
  fmt.Println("SendValues - publishing message...")
  return mqttClient.Publish("devices/"+uuid+"/values", qos, false, messageJSON)
}

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
    // ATTENTION!!!
    // To use "InsecureSkipVerify: false" you need to connect to MQTT using the public domain
    InsecureSkipVerify: false,
    // Certificates = list of certs client sends to server.
    Certificates: []tls.Certificate{cert},
  }
}

func InitMqtt() {
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

  mqttClient = mqtt.NewClient(opts)
  if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
    panic(token.Error())
  }

  time.Sleep(6 * time.Second)
}
