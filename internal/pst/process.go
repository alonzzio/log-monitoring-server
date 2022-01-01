package pst

import (
	"context"
	"github.com/alonzzio/log-monitoring-server/internal/helpers"
	"log"
	"sync"
)

// InitPubSubProcess will initialise pub/sub
// Create a topic from env variable
// Initialise and run Publishers Fake services concurrently.
func (repo *Repository) InitPubSubProcess(publishers, serviceNamePoolSize uint, w *sync.WaitGroup, serviceConfig PublisherServiceConfig) {
	defer w.Done()
	ctx := context.Background()
	c, err := repo.NewPubSubClient(ctx, repo.App.Environments.PubSub.ProjectID)
	if err != nil {
		// if we encounter error we can't continue
		log.Fatal(err)
	}

	err = repo.CreateTopic(ctx, repo.App.Environments.PubSub.TopicID, c)
	if err != nil {
		// if we encounter error we can't continue
		log.Fatal(err)
	}
	// External service fake pool
	ServNamePool := repo.GenerateServicesPool(serviceNamePoolSize)

	var wg sync.WaitGroup

	wg.Add(int(publishers))

	for i := uint(0); i < publishers; i++ {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for {
				m := repo.GenerateRandomMessages(serviceConfig.PerBatch, ServNamePool)
				errP := repo.PublishBulkMessage(repo.App.Environments.PubSub.TopicID, m, c, serviceConfig)
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
