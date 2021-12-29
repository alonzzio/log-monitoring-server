/*
	Log Monitoring System ( LMS )
	This program is indented to demonstrate the functionalities only.
	Not fully focused on complete error handling in place.
	However, I'll try my best to cover error handling in place

	Copy right: Ratheesh Alon Rajan
*/

package main

import (
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"github.com/alonzzio/log-monitoring-server/internal/pst"
	"log"
	"os"
	"sync"
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

	//ctx := context.Background()

	err := run()
	if err != nil {
		log.Fatal(err)
	}

	// init repositories
	pstRepo := pst.NewRepo(&app)
	pst.NewHandlers(pstRepo)

	/* Starting new pub/sub fake server */
	grpcCon, err := pst.StartPubSubFakeServer(9001)
	defer grpcCon.Close()

	app.GrpcPubSubServer.Conn = grpcCon

	/* Init PubSub services.
	This process simulates multiple or n number of services publishing messages to the given topic.
	We can control n and its frequency via env file.
	As this is an external service, I assume it runs continuously.*/

	msgConf := pst.PublisherServiceConfig{
		Frequency: time.Duration(int(app.Environments.PubSub.MessageFrequency)) * time.Millisecond,
		PerBatch:  app.Environments.PubSub.MessageBatch,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go pst.Repo.InitPubSubProcess(app.Environments.PubSub.ServicePublishers, app.Environments.PubSub.ServiceNamePool, &wg, msgConf)

	go func(wg *sync.WaitGroup) {
		wg.Wait()
	}(&wg)

	/*
		End of Publishing Services
	*/

	/*
		Start of Message Que and processing
	*/

	time.Sleep(100 * time.Second)
	fmt.Println("Shutting down Service!")
}
