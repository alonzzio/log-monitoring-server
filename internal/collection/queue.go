package collection

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"google.golang.org/api/option"
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

// Worker receives the message from pub/sub
// send message to 'results' channel
func (repo *Repository) Worker(wg *sync.WaitGroup) {
	defer wg.Done()
	ctx := context.Background()
	con := repo.App.GrpcPubSubServer.Conn
	client, err := pubsub.NewClient(ctx, repo.App.Environments.PubSub.ProjectID, option.WithGRPCConn(con))
	if err != nil {
		log.Println("Error: in client:", err)
	}
	defer client.Close()

	for {
		select {

		case _ = <-repo.Jobs:
			var mu sync.Mutex
			sub := client.Subscription(repo.App.Environments.PubSub.SubscriptionID)
			received := 0
			cctx, cancel := context.WithCancel(ctx)
			errR := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
				mu.Lock()
				defer mu.Unlock()
				msg.Ack()
				repo.Results <- Result{Data: msg.Data}
				received++
				if received == 10 {
					cancel()
				}
			})
			if errR != nil {
				fmt.Println("Err in receive:", err)
				continue
			}
		}
	}
}

// CreateWorkerPool creates worker pool
func (repo *Repository) CreateWorkerPool(noOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go repo.Worker(&wg)
	}

	wg.Wait()
}

// CreateMessageWorkerPool creates pool for processing messages from 'results' channel
func (repo *Repository) CreateMessageWorkerPool(noOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go repo.MessageProcessWorker(&wg)
	}

	wg.Wait()
}

// Allocate allocates job channel
func (repo *Repository) Allocate() {
	defer close(repo.Jobs)
	for {
		if len(repo.Jobs) == 100 {
			continue
		}
		job := Job{}
		repo.Jobs <- job
	}
}

// MessageProcessWorker process the received messages from 'results' channel
func (repo *Repository) MessageProcessWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {

		case _ = <-repo.Results:
			//fmt.Println("here message received")
			//fmt.Println(string(data.Data))

		}
	}
}
