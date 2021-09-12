// Air Conditioner API
//
// Air Conditioner control system APIs.
//
//	Schemes: http
//  Host: localhost:3000
//	BasePath: /
//	Version: 1.0.0
//	Contact: Stefano Cappa <stefano.cappa.ks89@gmail.com> https://github.com/Ks89
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
// swagger:meta
package main

import (
	"air-conditioner/handlers"
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
	"time"
)

var devicesHandler *handlers.DevicesHandler

func init() {
	ctx := context.Background()
	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collectionHomes := client.Database("airConditionerDb").Collection("homes")

	devicesHandler = handlers.NewDevicesHandler(ctx, collectionHomes)
}

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func initMqtt() {
	mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker("tcp://192.168.1.71:1883").SetClientID("apiServer")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := c.Subscribe("topic/state", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	//
	//for i := 0; i < 5; i++ {
	//	text := fmt.Sprintf("this is msg #%d!", i)
	//	token := c.Publish("topic/state", 0, false, text)
	//	token.Wait()
	//}

	time.Sleep(6 * time.Second)

	//if token := c.Unsubscribe("topic/state"); token.Wait() && token.Error() != nil {
	//	fmt.Println(token.Error())
	//	os.Exit(1)
	//}
	//
	//c.Disconnect(250)
	//
	//time.Sleep(1 * time.Second)
}

func main() {
	router := gin.Default()

	//config := cors.DefaultConfig()
	//config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	//router.Use(cors.New(config))

	router.POST("/authorize", devicesHandler.AuthorizeDeviceHandler)

	initMqtt()

	err := router.Run(":8081")
	if err != nil {
		panic(err)
	}
}
