package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

var channelAmpq *amqp.Channel
var logger *zap.SugaredLogger

func failOnError(err error, msg string) {
	if err != nil {
		logger.Fatalf("%s: %s", msg, err)
	}
}

func Publish(uuid string, payload []byte) {
	logger.Debug("Publishing payload with uuid: ", uuid)
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
		logger.Errorf("InitAmqpPublisher - Error conn: %v\n", err)
		defer conn.Close()
		return
	}

	channelAmpq, err = conn.Channel()
	failOnError(err, "Failed to open a channel")
	//defer channelAmpq.Close()
	if err != nil {
		logger.Errorf("InitAmqpPublisher - Error channelAmpq: %v\n", err)
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
		logger.Errorf("InitAmqpPublisher - Error queue declare: %v\n", err)
		return
	}
}
