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
	"log"
	"net/http"
	"os"
	"runtime"
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
	}(2500 * time.Minute)

	err := run()
	if err != nil {
		log.Fatal(err)
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

		log.Println(fmt.Sprintf("Data Access Server Started at port: %v ", app.Environments.DataAccessLayer.PortNumber))
		log.Fatal(srv.ListenAndServe())
	}(&wg)

	/*
		Data Collection Layer
	*/
	go func() {

		for {
			fmt.Println("Number of Go-routines:", runtime.NumGoroutine())
			time.Sleep(5000 * time.Millisecond)
		}
	}()

	c, err := pst.Repo.NewPubSubClient(context.Background(), app.Environments.PubSub.ProjectID)
	if err != nil {
		log.Fatal("Client creation err:", err)
	}

	_, err = pst.Repo.CreateSubscription(context.Background(), app.Environments.PubSub.SubscriptionID, app.Environments.PubSub.TopicID, c)
	if err != nil {
		log.Fatal("Subscription creation err:", err)
	}

	//go collection.Repo.Allocate()
	//go collection.Repo.CreateWorkerPool(app.Environments.DataCollectionLayer.Workers)
	//go collection.Repo.CreateMessageWorkerPool(app.Environments.DataCollectionLayer.Workers)

	//go func(wg *sync.WaitGroup) {
	//	wg.Wait()
	//}(&wg)

	jobs := make(chan collection.ReceiverJob, app.Environments.DataCollectionLayer.JobsBuffer)
	results := make(chan collection.ReceiverResult, app.Environments.DataCollectionLayer.ResultBuffer)
	logs := make(chan collection.Logs, app.Environments.DataCollectionLayer.LogsBuffer)

	var wgg sync.WaitGroup

	// Workers CPU core
	//n := runtime.NumCPU()
	//if we want to change use Worker from DCL env

	numWorkers := app.Environments.DataCollectionLayer.Workers
	go collection.Repo.CreateReceiverWorkerPools(numWorkers, jobs, results, &wgg)
	go collection.Repo.CreateJobsPool(jobs)
	go collection.Repo.CreateProcessWorkerPools(numWorkers, results, logs, &wg)
	// make our channels for communicating work and results
	// spin up workers and use a sync.WaitGroup to indicate completion
	//fmt.Println(runtime.NumCPU())
	////for i := 0; i < runtime.NumCPU; i++ {
	//for i := 0; i < runtime.NumCPU(); i++ {
	//	wgg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		collection.Repo.ReceiverWorker(jobs, results)
	//	}()
	//}
	// wait on the workers to finish and close the result channel
	// to signal downstream that all work is done
	//// start sending jobs
	//go func() {
	//	defer close(jobs)
	//	for {
	//		jobs <- collection.ReceiverJob{} // I haven't defined getJob() and noMoreJobs()
	//		//if noMoreJobs() {  // they are just for illustration
	//		//	break
	//		//}
	//	}
	//}()

	// read all the results
	func() {
		for {
			select {
			case l := <-logs:
				fmt.Println("Revd: len:", len(*l))
			}
		}
	}()

	time.Sleep(2500 * time.Second)
}
