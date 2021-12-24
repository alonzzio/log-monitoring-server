package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/option"
	"log"
	"os"
	"time"
)

type App struct {
	context context.Context
	client  *pubsub.Client
	config  Config
}

type Config struct {
	context          context.Context
	gcpProjectName   string
	subscriptionName string
	topicName        string
	options          []option.ClientOption
}

func newApp(config Config) (*App, error) {
	client, err := pubsub.NewClient(config.context, config.gcpProjectName, config.options...)
	if err != nil {
		return nil, err
	}

	return &App{context: config.context, client: client, config: config}, nil
}

func PubSubProcessing() {
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	psApp, err := newApp(Config{context: ctx, gcpProjectName: "test-lms", subscriptionName: "lms", topicName: "lms-topic", options: []option.ClientOption{option.WithoutAuthentication()}})
	if err != nil {
		log.Fatal(err)
	}
	prepare(psApp)

	//subscribe first
	func() {
		psApp.client.Topic("lms-topic").Publish(ctx, &pubsub.Message{Data: []byte("{\"greeting\" : \"hello\"}")})
		psApp.client.Subscription("lms").Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			var jsonMessage map[string]interface{}
			json.Unmarshal(msg.Data, &jsonMessage)
			fmt.Println("message: ", jsonMessage)
		})
	}()

	psApp.psRun()

}

func prepare(app *App) {
	//no error checking, since this is just a demo
	topic, _ := app.client.CreateTopic(app.config.context, "lms-topic")
	app.client.CreateSubscription(app.config.context, "lms", pubsub.SubscriptionConfig{Topic: topic})
	res, _ := app.client.CreateTopic(app.config.context, "lms")
	app.client.CreateSubscription(app.config.context, "lms", pubsub.SubscriptionConfig{Topic: res})
}

func (app *App) psRun() {

	log.Println("waiting for messages")

	app.client.Subscription(app.config.subscriptionName).Receive(app.config.context, func(ctx context.Context, message *pubsub.Message) {

		var messageJson map[string]interface{}

		json.Unmarshal(message.Data, &messageJson)

		log.Printf("received message with id: %s and content %v", message.ID, messageJson)

		messageJson["processed_time"] = time.Now()

		result, _ := json.Marshal(messageJson)

		app.client.Topic(app.config.topicName).Publish(ctx, &pubsub.Message{Data: result})

		message.Ack()
	})

	log.Println("stopped waiting for messages")
}
