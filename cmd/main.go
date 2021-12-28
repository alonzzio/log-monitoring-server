/*
	Log Monitoring System ( LMS )
	This program is indented to demonstrate the functionalities only.
	Not fully focused on complete error handling in place.
	However, I'll try my best to cover error handling in place

	Copy right: Ratheesh Alon Rajan
*/

package main

import (
	"context"
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"github.com/alonzzio/log-monitoring-server/internal/pst"
	"log"
	"os"
	"sync"

	"github.com/alonzzio/log-monitoring-server/internal/helpers"
	"time"
)

// app holds application wide configs
var app config.AppConfig

//var conn *config.Conn

func main() {
	// This go routine will shut down entire process after given duration
	go func(d time.Duration) {
		// Sleeps until this time then exits
		// Only for this exercise
		time.Sleep(d)
		log.Println("Shutting down Service...")
		os.Exit(0)
	}(65 * time.Second)

	ctx := context.Background()

	err := run()
	if err != nil {
		log.Fatal(err)
	}

	// init repositories
	pstRepo := pst.NewRepo(&app)
	pst.NewHandlers(pstRepo)

	// starting new pub/sub fake server
	grpcCon, err := pst.StartPubSubFakeServer(9001)
	defer grpcCon.Close()

	app.GrpcPubSubServer.Conn = grpcCon

	c, err := pst.Repo.NewPubSubClient(ctx, app.Environments.PubSub.ProjectID)

	err = pst.Repo.CreateTopic(ctx, app.Environments.PubSub.TopicID, c)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("topic")

	servNamePool := pst.Repo.GenerateServicesPool(10)
	for i := 0; i < 10; i++ {
		fmt.Println(pst.Repo.GetRandomServiceName(servNamePool))
	}
	err = pst.Repo.PublishBulkMessage(app.Environments.PubSub.TopicID, pst.Repo.GenerateRandomMessage(100, servNamePool), c)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("published")

	//
	/////// this point we have succesfully created message to pub sub.
	//// need to do it in loop and make it big but can be done later
	//
	//// now start subscribing
	//
	//subscriberClient, err := pubsub.NewClient(ctx, "project", option.WithGRPCConn(grpcConn))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//t := subscriberClient.Topic("lms-topic")
	//
	////subs, err := subscriberClient.CreateSubscription(ctx, "lms-sub", pubsub.SubscriptionConfig{Topic: t,
	////	AckDeadline:      10 * time.Second,
	////	ExpirationPolicy: 25 * time.Hour})
	//
	//_, err = subscriberClient.CreateSubscription(context.Background(), "lms-topic",
	//	pubsub.SubscriptionConfig{Topic: t})
	//
	//fmt.Println("reached here")
	//
	//subs := subscriberClient.Subscription("lms-topic")
	//fmt.Println(subs.ID())
	//
	//ok, err := subs.Exists(context.Background())
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println("subs exist:", ok)
	//
	//fmt.Println(subs.String())
	//err = subs.Receive(context.Background(),
	//	func(ctx context.Context, mm *pubsub.Message) {
	//		log.Printf("Got message: %s", mm.Data)
	//		mm.Ack()
	//	})
	//if err != nil {
	//	// Handle error.
	//	log.Fatal(err)
	//}
	for ii := 0; ii < 10; ii++ {
		go func() {
			fmt.Println("Goroutine ID:", helpers.GetGoRoutineID())
		}()
	}

	time.Sleep(1 * time.Minute)
	fmt.Println("Shutting down Service!")
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
