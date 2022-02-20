package amqp

import (
  "api-server/ws"
  "encoding/json"
  amqp "github.com/rabbitmq/amqp091-go"
  "go.uber.org/zap"
  "os"
)

var channelAmpq *amqp.Channel
var logger *zap.SugaredLogger

func InitAmqpSubscriber(log *zap.SugaredLogger) {
  logger = log
  rabbitMqURL := os.Getenv("RABBITMQ_URL")
  conn, err := amqp.Dial(rabbitMqURL)
  if err != nil {
    logger.Fatalf("Failed to connect to RabbitMQ: %s", err)
    defer conn.Close()
    return
  }

  channelAmpq, err = conn.Channel()
  if err != nil {
    logger.Fatalf("Failed to open a channel: %s", err)
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

  if err != nil {
    logger.Fatalf("Failed to declare a queue: %s", err)
    return
  }

  go Subscribe()
}

func Subscribe() {
  logger.Info("Subscribed to AMPQ")
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
    logger.Errorf("cannot consume from: %q, %v", "device", err)
    return
  }

  logger.Info("Subscribed to AMQP")

  for msg := range deliveries {
    logger.Info("msg.Body plain", msg.Body)
    var p2 interface{}
    err = json.Unmarshal(msg.Body, &p2)
    if err != nil {
      logger.Errorf("[Gin-OAuth] Failed to unmarshal client credentials: %v\n", err)
    }
    m := p2.(map[string]interface{})
    logger.Info(m)

    go ws.Send()

    // TODO handle this in some ways, because I have to process this data and send it via websocket to the front end
  }
}
