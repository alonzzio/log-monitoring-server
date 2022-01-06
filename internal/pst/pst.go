package pst

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"github.com/brianvoe/gofakeit/v6"
	"google.golang.org/api/option"
	"math/rand"
	"time"
)

type Severity int

const (
	Debug Severity = iota
	Info
	Warn
	Error
	Fatal
)

// Message holds the message structure
type Message struct {
	ServiceName string    `json:"service_name"`
	Payload     string    `json:"payload"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}

// Repository holds App config
type Repository struct {
	App *config.AppConfig
}

// PublisherServiceConfig holds the publisher configuration for service workers
type PublisherServiceConfig struct {
	Frequency time.Duration
	PerBatch  uint
}

// NewRepo initialise and return Repository Type Which holds AppConfig
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers  sets the repository  for the handlers
func NewHandlers(r *Repository) {
	Repo = r
	//logger = r.App.Logger.Logger
}

var Repo *Repository

//var logger zerolog.Logger

// GetPayLoad generates payload as a paragraph.
// Word count and Sentence count can be adjusted in env
// Do not set higher values, it will generate very long paragraphs.
// it can be problematic for SQL inserts and performance
func (repo *Repository) GetPayLoad() string {
	return gofakeit.Paragraph(1, repo.App.Environments.Paragraph.SentenceCount, repo.App.Environments.Paragraph.WordCount, ".")
}

// GetRandomSeverity generates random severity between the range
func (repo *Repository) GetRandomSeverity(min, max int) Severity {
	rand.Seed(time.Now().UnixNano())
	return Severity(rand.Intn(max-min+1) + min)
}

// GetRandomServiceName generates random service name for the message
// this function generates random string only
func (repo *Repository) GetRandomServiceName(s *[]string) string {
	min := 0
	max := len(*s) - 1
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(max-min+1) + min
	v := (*s)[i]
	return v
}

//// PublishMessage publishes a message to given topic
//func (repo *Repository) PublishMessage(topic string, m Message, c *pubsub.Client) error {
//	t := c.Topic(topic)
//	ctx := context.Background()
//	defer t.Stop()
//	var results []*pubsub.PublishResult
//	pr := t.Publish(ctx, &pubsub.Message{Data: []byte(fmt.Sprintf("%v", m))})
//	results = append(results, pr)
//	for _, rr := range results {
//		id, err := rr.Get(ctx)
//		if err != nil {
//			return err
//		}
//		fmt.Printf("Published a message with a message ID: %s\n", id)
//	}
//	return nil
//}

// PublishBulkMessage publishes a message to given topic
func (repo *Repository) PublishBulkMessage(topic string, msg *[]Message, c *pubsub.Client, msgConfig PublisherServiceConfig) error {
	t := c.Topic(topic)
	ctx := context.Background()
	defer t.Stop()
	for _, m := range *msg {
		var results []*pubsub.PublishResult
		out, err := json.Marshal(m)
		if err != nil {
			return err
		}

		pr := t.Publish(ctx, &pubsub.Message{Data: out})
		results = append(results, pr)
		for _, rr := range results {
			_, errGet := rr.Get(ctx) // _ is id
			if errGet != nil {
				return errGet
			}
			//fmt.Printf("Published a message with a message ID: %s\n", id)
		}
		time.Sleep(msgConfig.Frequency)
	}
	return nil
}

//// PublishBulkMessage publishes a message to given topic
//func (repo *Repository) PublishBulkMessage(topic string, msg *[]Message, c *pubsub.Client, msgConfig PublisherServiceConfig) error {
//	t := c.Topic(topic)
//	ctx := context.Background()
//	defer t.Stop()
//	for _, m := range *msg {
//		var results []*pubsub.PublishResult
//		out, err := json.Marshal(m)
//		if err != nil {
//			return err
//		}
//
//		pr := t.Publish(ctx, &pubsub.Message{Data: out})
//		results = append(results, pr)
//		for _, rr := range results {
//			_, errGet := rr.Get(ctx) // _ is id
//			if errGet != nil {
//				return errGet
//			}
//			//fmt.Printf("Published a message with a message ID: %s\n", id)
//		}
//		time.Sleep(msgConfig.Frequency)
//	}
//	return nil
//}

// NewPubSubClient creates a new client connection for pub/sub
func (repo *Repository) NewPubSubClient(ctx context.Context, projectID string) (*pubsub.Client, error) {
	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(repo.App.GrpcPubSubServer.Conn))
	if err != nil {
		return nil, err
	}
	return client, nil
}

// CreateSubscription creates a subscription for a given client on a topic
func (repo *Repository) CreateSubscription(ctx context.Context, subID string, topicName string, c *pubsub.Client) (*pubsub.Subscription, error) {
	t := c.Topic(topicName)
	// config can be pull from env if wanted to.
	s, err := c.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{Topic: t,
		AckDeadline:      60 * time.Second,
		ExpirationPolicy: 1 * time.Hour})
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateTopic receives the message from pub sub
func (repo *Repository) CreateTopic(ctx context.Context, topic string, c *pubsub.Client) error {
	_, err := c.CreateTopic(ctx, topic)
	if err != nil {
		return err
	}
	return nil
}

// GenerateRandomMessages for the pub sub
// it creates multiple messages as slice
func (repo *Repository) GenerateRandomMessages(n uint, serviceNames *[]string) *[]Message {
	m := make([]Message, 0)
	for i := uint(0); i < n; i++ {
		//compose message
		a := Message{
			ServiceName: repo.GetRandomServiceName(serviceNames),
			Payload:     repo.GetPayLoad(),
			Severity:    repo.SeverityToString(repo.GetRandomSeverity(0, 4)),
			Timestamp:   time.Now(),
		}
		m = append(m, a)
	}
	return &m
}

// GenerateARandomMessage for the pub sub
func (repo *Repository) GenerateARandomMessage(serviceNames *[]string) *Message {
	//compose message
	return &Message{
		ServiceName: repo.GetRandomServiceName(serviceNames),
		Payload:     repo.GetPayLoad(),
		Severity:    repo.SeverityToString(repo.GetRandomSeverity(0, 4)),
		Timestamp:   time.Now(),
	}
}

// SeverityToString converts severity to string
func (repo *Repository) SeverityToString(s Severity) string {
	var m string
	switch s {

	case 0:
		m = "Debug"
	case 1:
		m = "Info"
	case 2:
		m = "Warn"
	case 3:
		m = "Error"
	case 4:
		m = "Fatal"
	}
	return m
}

// GenerateServicesPool generate some service name for this exercise
// This function generates "Service-name:1" "Service-name:2"...
func (repo *Repository) GenerateServicesPool(n uint) *[]string {
	var s []string
	for i := uint(0); i < n; i++ {
		s = append(s, fmt.Sprintf("Service-name:%v", i+1))
	}
	return &s
}
