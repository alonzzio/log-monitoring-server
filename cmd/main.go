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
	"github.com/alonzzio/log-monitoring-server/internal/lmslogging"
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
	// lmsLogChan chan receives logs from application/system.
	lmsLogChan := make(chan lmslogging.Log, 100)
	lmsLogs := lmslogging.LmsLogging{}
	lg, sysFile, appFile, err := lmsLogs.NewSysAndAppFileLog("../LMSlogs/system.log", "../LMSlogs/app.log")

	// app.log and system.log files
	defer sysFile.Close()
	defer appFile.Close()

	// initiate internal centralised log writing for app.log and system.log files
	go lg.LogWriter(lmsLogChan)

	// Write First System Log Message
	lmsLogChan <- lmslogging.Log{
		SysLog:   true,
		Severity: lmslogging.Info,
		Prefix:   "LMS",
		Message:  "<============ Log Monitoring Server ============>",
	}
	// Write First app Log Message
	lmsLogChan <- lmslogging.Log{
		SysLog:   false,
		Severity: lmslogging.Info,
		Prefix:   "LMS",
		Message:  "<============ Log Monitoring Server ============>",
	}

	err = run(lmsLogChan)
	if err != nil {
		lmsLogChan <- lmslogging.Log{
			SysLog:   true,
			Severity: lmslogging.Fatal,
			Prefix:   "AppInitRun",
			Message:  err.Error(),
		}
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
	go pst.Repo.InitPubSubProcess(app.Environments.PubSub.ServicePublishers, app.Environments.PubSub.ServiceNamePool, lmsLogChan, &wg, msgConf)

	/* Data Access Layer */
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		srv := &http.Server{
			Addr:         fmt.Sprintf(":%v", app.Environments.DataAccessLayer.PortNumber),
			Handler:      routes(),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		lmsLogChan <- lmslogging.Log{
			SysLog:   true,
			Severity: lmslogging.Info,
			Prefix:   "DataAccessLayer",
			Message:  "Data Access Server Started at port:" + string(app.Environments.DataAccessLayer.PortNumber),
		}
		log.Fatal(srv.ListenAndServe())
	}(&wg)

	/* Data Collection Layer */
	c, err := pst.Repo.NewPubSubClient(context.Background(), app.Environments.PubSub.ProjectID)
	if err != nil {
		lmsLogChan <- lmslogging.Log{
			SysLog:   true,
			Severity: lmslogging.Fatal,
			Prefix:   "Publisher",
			Message:  err.Error(),
		}
		log.Fatal(err)
	}

	_, err = pst.Repo.CreateSubscription(context.Background(), app.Environments.PubSub.SubscriptionID, app.Environments.PubSub.TopicID, c)
	if err != nil {
		lmsLogChan <- lmslogging.Log{
			SysLog:   true,
			Severity: lmslogging.Fatal,
			Prefix:   "Publisher",
			Message:  err.Error(),
		}
		log.Fatal(err)
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
	go collection.Repo.CreateReceiverWorkerPools(numWorkers, jobs, results, lmsLogChan, &wgg)
	go collection.Repo.CreateJobsPool(jobs)
	go collection.Repo.CreateProcessWorkerPools(numWorkers, results, logsBatch, &wg)
	go collection.Repo.CreateDbProcessWorkerPools(numWorkers, logsBatch, logsBatch, lmsLogChan, &wg)

	// This go routine will shut down entire process after given duration
	// Sleeps until this time then exits
	// Only for this exercise
	go func(d time.Duration, lmsLogChan chan<- lmslogging.Log) {
		time.Sleep(d)
		fmt.Println("Shutting down Service...")
		lmsLogChan <- lmslogging.Log{
			SysLog:   false,
			Severity: lmslogging.Fatal,
			Prefix:   "AppShutDown",
			Message:  "<========= Shutting down Service... =========>",
		}
		// Let logs write its final message
		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}(2000*time.Second, lmsLogChan)

	go func() {
		for {
			fmt.Println("No of Go Routines:", runtime.NumGoroutine())
			time.Sleep(1 * time.Second)
		}

	}()

	fmt.Println("Log Monitoring Server Started.")
	time.Sleep(2000 * time.Second)
}
