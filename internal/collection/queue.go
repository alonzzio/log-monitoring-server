package collection

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/pst"
	"log"
	"sync"
	"time"
)

// Message holds the message structure
type Message struct {
	ServiceName string    `json:"service_name"`
	Payload     string    `json:"payload"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}

type Job struct{}

type Result struct {
	Data []byte
}

var jobs = make(chan Job, 100)
var results = make(chan Result, 100)

func Worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case _ = <-jobs:
			ctx := context.Background()
			client, err := pst.Repo.NewPubSubClient(ctx, "lms")
			if err != nil {
				log.Println("Error: in client:", err)
				return
			}
			defer client.Close()
			sub := client.Subscription("lms-sub")
			var mu sync.Mutex
			received := 0
			cctx, cancel := context.WithCancel(ctx)
			err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
				mu.Lock()
				defer mu.Unlock()
				msg.Ack()
				results <- Result{Data: msg.Data}
				received++
				if received == 10 {
					cancel()
				}
			})
			if err != nil {
				fmt.Println("Err in receive", err)
			}
		}
	}
}

func CreateWorkerPool(noOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go Worker(&wg)
	}

	wg.Wait()
	fmt.Println("Closed Worker")
	//close(results)
}

func Allocate(wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(jobs)
	for {
		if len(jobs) == 1000 {
			continue
		}
		job := Job{}
		jobs <- job
	}
}

func ResultFunc(wg *sync.WaitGroup) {
	defer wg.Done()
	for result := range results {
		fmt.Println(string(result.Data))
	}

}
