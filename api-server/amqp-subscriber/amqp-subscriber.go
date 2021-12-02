package amqp

import (
	"api-server/ws"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

var channelAmpq *amqp.Channel

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func Subscribe() {
	fmt.Println("Subscribed to AMPQ")
	deliveries, err := channelAmpq.Consume(
		"device",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Printf("cannot consume from: %q, %v", "device", err)
		return
	}

	log.Printf("subscribed...")

	for msg := range deliveries {
		fmt.Println("msg.Body plain", msg.Body)
		var p2 interface{}
		json.Unmarshal(msg.Body, &p2)
		m := p2.(map[string]interface{})
		fmt.Println(m)

		go ws.Send()
		//go ws.Send(msg.Body)
		// TODO handle this in some ways, because I have to process this data and send it via websocket to the front end
	}
}

func InitAmqpSubscriber() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	if err != nil {
		fmt.Println("Error conn: ", err)
		defer conn.Close()
		return
	}

	channelAmpq, err = conn.Channel()
	failOnError(err, "Failed to open a channel")
	//defer channelAmpq.Close()
	if err != nil {
		fmt.Println("Error channelAmpq: ", err)
		defer channelAmpq.Close()
		return
	}

	_, err = channelAmpq.QueueDeclare(
		"device",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")
	if err != nil {
		fmt.Println("Error queue declare: ", err)
		return
	}

	go Subscribe()
}
