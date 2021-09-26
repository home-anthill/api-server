package mqtt_client

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

const qos byte = 0

var c mqtt.Client

var defaultHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("---UNKNOWN TOPIC---")
	fmt.Printf("MessageID: %s\n", msg.MessageID())
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
func SendFan(uuid string, messageJSON []byte) mqtt.Token {
	fmt.Println("SendFan - publishing message...")
	return c.Publish("devices/"+uuid+"/fan", qos, false, messageJSON)
}
func SendSwing(uuid string, messageJSON []byte) mqtt.Token {
	fmt.Println("SendSwing - publishing message...")
	return c.Publish("devices/"+uuid+"/swing", qos, false, messageJSON)
}

func InitMqtt() {
	//mqtt.DEBUG = log.New(os.Stdout, "", 0)
	//mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker("tcp://192.168.1.71:1883").SetClientID("apiServer")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(defaultHandler)
	opts.SetPingTimeout(1 * time.Second)

	c = mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	c.Subscribe("devices/*/onoff", qos, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("MessageID: %s\n", msg.MessageID())
		fmt.Printf("Topic: %s\n", msg.Topic())
		fmt.Printf("Payload: %s\n", msg.Payload())
		// TODO receive changes from devices and update api-server/apps and so on in next releases
	})
	c.Subscribe("devices/*/temperature", qos, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("MessageID: %s\n", msg.MessageID())
		fmt.Printf("Topic: %s\n", msg.Topic())
		fmt.Printf("Payload: %s\n", msg.Payload())
		// TODO receive changes from devices and update api-server/apps and so on in next releases
	})
	c.Subscribe("devices/*/mode", qos, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("MessageID: %s\n", msg.MessageID())
		fmt.Printf("Topic: %s\n", msg.Topic())
		fmt.Printf("Payload: %s\n", msg.Payload())
		// TODO receive changes from devices and update api-server/apps and so on in next releases
	})
	c.Subscribe("devices/*/fan", qos, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("MessageID: %s\n", msg.MessageID())
		fmt.Printf("Topic: %s\n", msg.Topic())
		fmt.Printf("Payload: %s\n", msg.Payload())
		// TODO receive changes from devices and update api-server/apps and so on in next releases
	})
	c.Subscribe("devices/*/swing", qos, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("MessageID: %s\n", msg.MessageID())
		fmt.Printf("Topic: %s\n", msg.Topic())
		fmt.Printf("Payload: %s\n", msg.Payload())
		// TODO receive changes from devices and update api-server/apps and so on in next releases
	})

	time.Sleep(6 * time.Second)

	//c.Disconnect(250)
	//time.Sleep(1 * time.Second)
}
