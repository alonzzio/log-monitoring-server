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
	"time"
)

// app holds application wide configs
var app config.AppConfig

func main() {
	// This go routine will shut down entire process after given duration
	// Sleeps until this time then exits
	// Only for this exercise
	go func(d time.Duration) {
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

	fmt.Println("Shutting down Service!")
}
