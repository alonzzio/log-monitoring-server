/*
	Log Monitoring System ( LMS )
	This program is indented to demonstrate the functionalities only.
	Not fully focused on complete error handling in place.
	However, I'll try my best to cover error handling in place

	Copy right: Ratheesh Alon Rajan
*/

package main

import (
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"context"
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"log"
	"os"
	"sync"
	"time"
)

// app holds application wide configs
var app config.AppConfig

func main() {
	ctx := context.Background()
	err := run()
	if err != nil {
		log.Fatal(err)
	}

	var conn *config.Conn
	conn, err = newConn()
	if err != nil {
		log.Fatal(err)
	}

	// set conn to App
	// When we reach this point, successful mysql/any other Db connection is ready to use
	app.Conn = conn

	err = initialiseDatabase(&app)
	if err != nil {
		log.Fatal(err)
	}

	grpcConn, psServer, err := pubsubFakeServer("lms-topic", ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer psServer.Close()
	defer grpcConn.Close()

	client, err := pubsub.NewClient(ctx, "project", option.WithGRPCConn(grpcConn))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		m := Message{
			ServiceName: "service1",
			Payload:     "dfgdfgfd dfgdfgdfg dfgdfhfgjfg fsdfsgdfh efrsghfghfgjgf ",
			Severity:    "INFO",
			Timestamp:   time.Now(),
		}

		err = publishMessage("lms-topic", m, client)
		if err != nil {
			log.Fatal(err)
		}

		m = Message{
			ServiceName: "service1",
			Payload:     "dfgdfdfsfds thsis sisaf  the besr gfd dfgdfgdfg dfgdfhfgjfg fsdfsgdfh efrsghfghfgjgf ",
			Severity:    "Err",
			Timestamp:   time.Now(),
		}

		err = publishMessage("lms-topic", m, client)
		if err != nil {
			log.Fatal(err)
		}

		m = Message{
			ServiceName: "service1",
			Payload:     "dfgdfdfsfds thdsggdsgd fhgcfhfhcvbbc sis sisaf  the besr gfd dfgdfgdfg dfgdfhfgjfg fsdfsgdfh efrsghfghfgjgf ",
			Severity:    "WARN",
			Timestamp:   time.Now(),
		}

		err = publishMessage("lms-topic", m, client)
		if err != nil {
			log.Fatal(err)
		}
		m = Message{
			ServiceName: "service1",
			Payload:     "dfgdfdfsfdxcvxcvvcxs thsis sisaf  the besr gfd dfgdfgdfg dfgdfhfgjfg fsdfsgdfh efrsghfghfgjgf ",
			Severity:    "INFO",
			Timestamp:   time.Now(),
		}

		err = publishMessage("lms-topic", m, client)
		if err != nil {
			log.Fatal(err)
		}
	}()

	///// this point we have succesfully created message to pub sub.
	// need to do it in loop and make it big but can be done later

	// now start subscribing

	subscriberClient, err := pubsub.NewClient(ctx, "project", option.WithGRPCConn(grpcConn))
	if err != nil {
		log.Fatal(err)
	}

	t := subscriberClient.Topic("lms-topic")

	//subs, err := subscriberClient.CreateSubscription(ctx, "lms-sub", pubsub.SubscriptionConfig{Topic: t,
	//	AckDeadline:      10 * time.Second,
	//	ExpirationPolicy: 25 * time.Hour})

	_, err = subscriberClient.CreateSubscription(context.Background(), "lms-topic",
		pubsub.SubscriptionConfig{Topic: t})

	fmt.Println("reached here")

	subs := subscriberClient.Subscription("lms-topic")
	fmt.Println(subs.ID())

	ok, err := subs.Exists(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("subs exist:", ok)

	fmt.Println(subs.String())
	err = subs.Receive(context.Background(),
		func(ctx context.Context, mm *pubsub.Message) {
			log.Printf("Got message: %s", mm.Data)
			mm.Ack()
		})
	if err != nil {
		// Handle error.
		log.Fatal(err)
	}

	fmt.Println("end")
}

func initialiseDatabase(app *config.AppConfig) error {
	_, err := app.Conn.DB.Exec(`CREATE DATABASE IF NOT EXISTS ` + os.Getenv("MYSQLDBNAME") + `;`)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	wg.Add(2)
	errChan := make(chan error, 2)

	sql1 := `CREATE TABLE IF NOT EXISTS service_logs (
			service_name VARCHAR(100) NOT NULL,
			payload VARCHAR(2048) NOT NULL,
			severity ENUM("debug", "info", "warn", "error", "fatal") NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
			);`

	sql2 := `CREATE TABLE IF NOT EXISTS service_severity (
			service_name VARCHAR(100) NOT NULL,
			severity ENUM("debug", "info", "warn", "error", "fatal") NOT NULL,
			count INT(4) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
			);`

	go executeSQLWorker(sql1, app, errChan, &wg)
	go executeSQLWorker(sql2, app, errChan, &wg)

	wg.Wait()
	close(errChan)

	for err = range errChan {
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

// executeSQLWorker this function executes against DB and passing errors through error channel
// this is not really needed but i am demonstrating the sql can be run parallel
func executeSQLWorker(sql string, app *config.AppConfig, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	// ignoring the result part here
	_, err := app.Conn.DB.Exec(sql)
	if err != nil {
		errChan <- err
	}

	errChan <- nil
}

// pubsubFakeServer startup a fake server for pub sub
func pubsubFakeServer(topic string, ctx context.Context) (grpcConn *grpc.ClientConn, psServer *pstest.Server, err error) {
	// Start a fake server running locally at 9001.
	srv := pstest.NewServerWithPort(9001)
	//defer srv.Close()
	// Connect to the server without using TLS.
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}

	log.Println("Pub Sub Server Started at port:9001")

	//defer conn.Close()
	// Use the connection when creating a pubsub client.

	client, err := pubsub.NewClient(ctx, "project", option.WithGRPCConn(conn))
	if err != nil {
		return nil, nil, err
	}
	_, err = client.CreateTopic(ctx, topic)
	if err != nil {
		return nil, nil, err
	}
	log.Println("Topic Created: ", topic)
	return conn, srv, nil

}

// createTopic creates topic on pub sub
func createTopic(topicName string, ctx context.Context, client *pubsub.Client) error {
	fmt.Println("came to topic")
	topic, err := client.CreateTopic(ctx, topicName)
	if err != nil {
		return err
	}
	log.Println("Topic Created: ", topic)
	return nil
}

func publishMessage(topic string, m Message, c *pubsub.Client) error {
	t := c.Topic(topic)
	ctx := context.Background()
	defer t.Stop()
	var results []*pubsub.PublishResult
	r := t.Publish(ctx, &pubsub.Message{Data: []byte(fmt.Sprintf("%v", m))})

	results = append(results, r)
	for _, r := range results {
		id, err := r.Get(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Published a message with a message ID: %s\n", id)
	}
	return nil
}

type Message struct {
	ServiceName string    `json:"service_name"`
	Payload     string    `json:"payload"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}
