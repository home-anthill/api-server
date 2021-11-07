package amqp

import (
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

func Publish(uuid string, payload []byte) {
	fmt.Println("Publishing payload with uuid: ", uuid)
	err := channelAmpq.Publish(
		"",
		"ac",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		})
	failOnError(err, "Failed to publish a message")
}

func InitAmqpPublisher() {
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
		"ac",
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
}
