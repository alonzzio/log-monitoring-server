package pst

import (
	"context"
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/helpers"
	"log"
	"sync"
)

// InitPubSubProcess will initialise pub/sub
// Create a topic from env variable
// Initialise and run Publishers Fake services concurrently.
// TODO BUG FIX: Implement worker pool
func (r *Repository) InitPubSubProcess(publishers, serviceNamePoolSize uint, w *sync.WaitGroup, serviceConfig PublisherServiceConfig) {
	defer w.Done()
	ctx := context.Background()
	c, err := r.NewPubSubClient(ctx, r.App.Environments.PubSub.ProjectID)
	if err != nil {
		// if we encounter error we can't continue
		log.Fatal(err)
	}

	err = r.CreateTopic(ctx, r.App.Environments.PubSub.TopicID, c)
	if err != nil {
		// if we encounter error we can't continue
		log.Fatal(err)
	}
	// External service fake pool
	ServNamePool := r.GenerateServicesPool(serviceNamePoolSize)

	var wg sync.WaitGroup

	wg.Add(int(publishers))

	for i := uint(0); i < publishers; i++ {
		// This wg is just for continuing the process

		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			fmt.Println("New publisher service stated in goroutine id:", helpers.GetGoRoutineID())
			for {
				m := r.GenerateRandomMessages(serviceConfig.PerBatch, ServNamePool)
				errP := r.PublishBulkMessage(r.App.Environments.PubSub.TopicID, m, c, serviceConfig)
				if errP != nil {
					log.Println("Error occurred in loop in goroutine id:", helpers.GetGoRoutineID())
					continue
				}
			}
		}(&wg)

		go func() {
			wg.Wait()
		}()
	}
}
