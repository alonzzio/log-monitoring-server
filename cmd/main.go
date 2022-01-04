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
	"github.com/alonzzio/log-monitoring-server/internal/access"
	"github.com/alonzzio/log-monitoring-server/internal/collection"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"github.com/alonzzio/log-monitoring-server/internal/pst"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	//"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// app holds application wide configs
var app config.AppConfig
var logger zerolog.Logger

func main() {
	// This go routine will shut down entire process after given duration
	// Sleeps until this time then exits
	// Only for this exercise
	go func(d time.Duration) {
		time.Sleep(d)
		logger.Info().Msg("Shutting down Service...")
		//log.Println("Shutting down Service...")
		os.Exit(0)
	}(2 * time.Minute)

	//initialise logging
	// Delete old if exist
	err := os.Remove("../logs/logs.log")
	if err != nil && !os.IsNotExist(err) {
		log.Fatal().Msg(err.Error())
	}

	var f *os.File
	f, err = os.OpenFile("../logs/logs.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	defer f.Close()
	logger = zerolog.New(f).With().Timestamp().Logger()

	app.Logger.Logger = logger

	err = run()
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	// init repositories
	// pub/sub
	pstRepo := pst.NewRepo(&app)
	pst.NewHandlers(pstRepo)

	// Data collection Layer
	dclRepo := collection.NewRepo(&app)
	collection.NewHandlers(dclRepo)

	// Data access Layer
	dalRepo := access.NewRepo(&app)
	access.NewHandlers(dalRepo)

	/* Starting new pub/sub fake server */
	grpcCon, pubSubServer, err := pst.StartPubSubFakeServer(9001)
	defer grpcCon.Close()
	defer pubSubServer.Close()

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
	wg.Add(2)
	go pst.Repo.InitPubSubProcess(app.Environments.PubSub.ServicePublishers, app.Environments.PubSub.ServiceNamePool, &wg, msgConf)

	/*
		Data Access Layer
	*/
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		srv := &http.Server{
			Addr:         fmt.Sprintf(":%v", app.Environments.DataAccessLayer.PortNumber),
			Handler:      routes(),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		logger.Info().Msg("Data Access Server Started at port: %v " + app.Environments.DataAccessLayer.PortNumber)
		servError := srv.ListenAndServe()
		log.Fatal().Msg(servError.Error())
	}(&wg)

	// This function is just for monitoring the Go-routines Surge when Bigger number of workers in place
	// And also used to track for Data Race
	go func() {
		for {
			//fmt.Println("Number of go-routines:", runtime.NumGoroutine())
			//logger.Println("Number of go-routines:", runtime.NumGoroutine())
			//logger.Print("test")
			time.Sleep(10 * time.Second)
		}
	}()

	/*
		Data Collection Layer
	*/
	c, err := pst.Repo.NewPubSubClient(context.Background(), app.Environments.PubSub.ProjectID)
	if err != nil {
		//	log.Fatal("Client creation err:", err)
		logger.Fatal().Msg("Client creation err: " + err.Error())
	}

	_, err = pst.Repo.CreateSubscription(context.Background(), app.Environments.PubSub.SubscriptionID, app.Environments.PubSub.TopicID, c)
	if err != nil {
		//log.Fatal("Subscription creation err:", err)
		logger.Fatal().Msg("Subscription creation err: " + err.Error())
	}

	go func(wg *sync.WaitGroup) {
		wg.Wait()
	}(&wg)

	jobs := make(chan collection.ReceiverJob, app.Environments.DataCollectionLayer.JobsBuffer)
	results := make(chan collection.ReceiverResult, app.Environments.DataCollectionLayer.ResultBuffer)
	logsBatch := make(chan collection.LogsBatch, app.Environments.DataCollectionLayer.LogsBuffer)

	var wgg sync.WaitGroup

	// Workers CPU core
	//n := runtime.NumCPU()
	//if we want to change use Worker from DCL env
	numWorkers := app.Environments.DataCollectionLayer.Workers
	go collection.Repo.CreateReceiverWorkerPools(numWorkers, jobs, results, &wgg)
	go collection.Repo.CreateJobsPool(jobs)
	go collection.Repo.CreateProcessWorkerPools(numWorkers, results, logsBatch, &wg)
	go collection.Repo.CreateDbProcessWorkerPools(numWorkers, logsBatch, logsBatch, &wg)

	fmt.Println("Log Monitoring Server Started.")
	time.Sleep(2 * time.Minute)
}
