package collection

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/pst"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

type Job struct {
	//Subscription *pubsub.Subscription
}

type Result struct {
	job  Job
	Data []byte
}

var jobs = make(chan Job, 100)
var results = make(chan Result, 100)

func Worker(wg *sync.WaitGroup) {
	defer wg.Done()

	ctx := context.Background()
	client, err := pst.Repo.NewPubSubClient(ctx, "lms")
	if err != nil {
		log.Println("Error: in client:", err)
		//continue
	}

	defer client.Close()

	sub := client.Subscription("lms-sub")
	for job := range jobs {
		// Turn on synchronous mode. This makes the subscriber use the Pull RPC rather
		// than the StreamingPull RPC, which is useful for guaranteeing MaxOutstandingMessages,
		// the max number of messages the client will hold in memory at a time.
		sub.ReceiveSettings.Synchronous = true
		sub.ReceiveSettings.MaxOutstandingMessages = 10

		// Receive messages for 10 seconds.
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		// Create a channel to handle messages to as they come in.
		cm := make(chan *pubsub.Message)
		defer close(cm)

		// Handle individual messages in a goroutine.
		go func() {
			for msg := range cm {
				//fmt.Println("got message",msg.Data)
				msg.Ack()
				output := Result{job, msg.Data}
				results <- output
			}
		}()

		// Receive blocks until the passed in context is done.
		err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			cm <- msg
		})
		if err != nil && status.Code(err) != codes.Canceled {
			fmt.Println("Receive: err", err)
		}

		//output := Result{job,}
		//results <- output
	}

	wg.Done()
}
func CreateWorkerPool(noOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go Worker(&wg)
	}
	wg.Wait()
	close(results)
}

func Allocate(wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(jobs)
	for {
		job := Job{}
		jobs <- job
	}
}

func ResultFunc(wg *sync.WaitGroup) {
	defer wg.Done()
	for _ = range results {
		//fmt.Println(string(result.Data))
	}

}
