package collection

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"github.com/alonzzio/log-monitoring-server/internal/lmslogging"
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

type ServiceSeverity struct {
	ServiceName string `json:"service_name"`
	Severity    string `json:"severity"`
	Count       int    `json:"count"`
}

type ReceiverJob struct{}

type ReceiverResult struct {
	Data []byte
}

type Logs []Message

type LogsBatch struct {
	LogMessage      []Message
	ServiceSeverity []ServiceSeverity
}

// ReceiverWorker receives messages from pub/sub and send it to receiverResult Channel
func (repo *Repository) ReceiverWorker(jobs <-chan ReceiverJob, results chan<- ReceiverResult, logs chan<- lmslogging.Log) {
	ctx := context.Background()
	con := repo.App.GrpcPubSubServer.Conn
	client, err := pubsub.NewClient(ctx, repo.App.Environments.PubSub.ProjectID, option.WithGRPCConn(con))
	if err != nil {
		logs <- lmslogging.Log{
			SysLog:   true,
			Severity: lmslogging.Error,
			Prefix:   "Receiver",
			Message:  err.Error(),
		}
		return
	}
	defer client.Close()
	// pop out jobs
	for range jobs {
		var mu sync.Mutex
		sub := client.Subscription(repo.App.Environments.PubSub.SubscriptionID)
		received := 0
		cctx, cancel := context.WithCancel(ctx)
		errR := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
			mu.Lock()
			defer mu.Unlock()
			msg.Ack()
			results <- ReceiverResult{Data: msg.Data}
			received++
			if received == repo.App.Environments.DataCollectionLayer.MessagePerReceive {
				cancel()
			}
		})
		if errR != nil {
			logs <- lmslogging.Log{
				SysLog:   true,
				Severity: lmslogging.Error,
				Prefix:   "Receiver",
				Message:  err.Error(),
			}
			continue
		}
	}
}

// CreateJobsPool sending unlimited jobs to ReceiverJobs Channel
func (repo *Repository) CreateJobsPool(jobs chan<- ReceiverJob) {
	defer close(jobs) // never closes though
	for {
		jobs <- ReceiverJob{}
	}
}

// CreateReceiverWorkerPools creates a pool of Receiver Workers
func (repo *Repository) CreateReceiverWorkerPools(poolSize int, jobs <-chan ReceiverJob, results chan<- ReceiverResult, logs chan<- lmslogging.Log, wg *sync.WaitGroup) {
	wg.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		go func(jobs <-chan ReceiverJob, results chan<- ReceiverResult) {
			defer wg.Done()
			repo.ReceiverWorker(jobs, results, logs)

		}(jobs, results)
	}
	wg.Wait()
}

// CreateProcessWorkerPools creates a pool of Receiver Workers
func (repo *Repository) CreateProcessWorkerPools(poolSize int, results <-chan ReceiverResult, logsBatch chan<- LogsBatch, wg *sync.WaitGroup) {
	wg.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		// message size can be controlled through env files
		go repo.MessageProcessWorker(repo.App.Environments.DataCollectionLayer.MessageBatchSize, results, logsBatch)
	}
	wg.Wait()
}

// MessageProcessWorker gets the messages from results channel and process as batch send to 'Logs' channel
func (repo *Repository) MessageProcessWorker(msgSize int, results <-chan ReceiverResult, logsBatch chan<- LogsBatch) {
	for {
		var batch []Message
		var serviceSeverity []ServiceSeverity
		for i := 0; i < msgSize; i++ {
			out := <-results
			var m Message
			err := json.Unmarshal(out.Data, &m)
			if err != nil {
				log.Println(err)
				continue
			}

			ss := ServiceSeverity{
				ServiceName: m.ServiceName,
				Severity:    "",
				Count:       1,
			}

			switch m.Severity {
			case "Debug":
				if len(serviceSeverity) == 0 {
					ss.Severity = "Debug"
					serviceSeverity = append(serviceSeverity, ss)
					break
				}

				found := false
				for ii, v := range serviceSeverity {
					if v.ServiceName == m.ServiceName && v.Severity == m.Severity {
						serviceSeverity[ii].Count += 1
						found = true
						break
					}
				}
				if !found {
					ss.Severity = "Debug"
					serviceSeverity = append(serviceSeverity, ss)
				}
			case "Info":
				if len(serviceSeverity) == 0 {
					ss.Severity = "Info"
					serviceSeverity = append(serviceSeverity, ss)
					break
				}

				found := false
				for ii, v := range serviceSeverity {
					if v.ServiceName == m.ServiceName && v.Severity == m.Severity {
						serviceSeverity[ii].Count += 1
						found = true
						break
					}
				}
				if !found {
					ss.Severity = "Info"
					serviceSeverity = append(serviceSeverity, ss)
				}
			case "Warn":
				if len(serviceSeverity) == 0 {
					ss.Severity = "Warn"
					serviceSeverity = append(serviceSeverity, ss)
					break
				}

				found := false
				for ii, v := range serviceSeverity {
					if v.ServiceName == m.ServiceName && v.Severity == m.Severity {
						serviceSeverity[ii].Count += 1
						found = true
						break
					}
				}
				if !found {
					ss.Severity = "Warn"
					serviceSeverity = append(serviceSeverity, ss)
				}
			case "Error":
				if len(serviceSeverity) == 0 {
					ss.Severity = "Error"
					serviceSeverity = append(serviceSeverity, ss)
					break
				}

				found := false
				for ii, v := range serviceSeverity {
					if v.ServiceName == m.ServiceName && v.Severity == m.Severity {
						serviceSeverity[ii].Count += 1
						found = true
						break
					}
				}
				if !found {
					ss.Severity = "Error"
					serviceSeverity = append(serviceSeverity, ss)
				}
			case "Fatal":
				if len(serviceSeverity) == 0 {
					ss.Severity = "Fatal"
					serviceSeverity = append(serviceSeverity, ss)
					break
				}

				found := false
				for ii, v := range serviceSeverity {
					if v.ServiceName == m.ServiceName && v.Severity == m.Severity {
						serviceSeverity[ii].Count += 1
						found = true
						break
					}
				}
				if !found {
					ss.Severity = "Fatal"
					serviceSeverity = append(serviceSeverity, ss)
				}
			}
			batch = append(batch, m)
		}

		logsBatch <- LogsBatch{
			LogMessage:      batch,
			ServiceSeverity: serviceSeverity,
		}
	}
}
