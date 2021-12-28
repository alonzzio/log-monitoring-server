package pst

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"github.com/brianvoe/gofakeit/v6"
	"google.golang.org/api/option"
	"math/rand"
	"sync"
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

type ServiceConfig struct {
	PublishFrequency time.Duration
	PerBatch         uint64
	Mutex            *sync.Mutex
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
}

var Repo *Repository

// GetPayLoad generates payload as a paragraph.
// Word count and Sentence count can be adjusted in env
// Do not set higher values, it will generate very long paragraphs.
// it can be problematic for SQL inserts and performance
func (r *Repository) GetPayLoad() string {
	return gofakeit.Paragraph(1, r.App.Environments.SentenceCount, r.App.Environments.WordCount, ".")
}

// GetRandomSeverity generates random severity between the range
func (r *Repository) GetRandomSeverity(min, max int) Severity {
	rand.Seed(time.Now().UnixNano())
	return Severity(rand.Intn(max-min+1) + min)
}

// GetRandomServiceName generates random service name for the message
func (r *Repository) GetRandomServiceName(s *[]string) string {
	min := 0
	max := len(*s) - 1
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(max-min+1) + min
	v := (*s)[i]
	return v
}

// PublishMessage publishes a message to given topic
func (r *Repository) PublishMessage(topic string, m Message, c *pubsub.Client) error {
	t := c.Topic(topic)
	ctx := context.Background()
	defer t.Stop()
	var results []*pubsub.PublishResult
	pr := t.Publish(ctx, &pubsub.Message{Data: []byte(fmt.Sprintf("%v", m))})
	results = append(results, pr)
	for _, rr := range results {
		id, err := rr.Get(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Published a message with a message ID: %s\n", id)
	}
	return nil
}

// PublishBulkMessage publishes a message to given topic
func (r *Repository) PublishBulkMessage(topic string, msg []Message, c *pubsub.Client) error {
	t := c.Topic(topic)
	ctx := context.Background()
	defer t.Stop()
	for m := range msg {
		var results []*pubsub.PublishResult
		pr := t.Publish(ctx, &pubsub.Message{Data: []byte(fmt.Sprintf("%v", m))})
		results = append(results, pr)
		for _, rr := range results {
			_, err := rr.Get(ctx) // _ is id
			if err != nil {
				return err
			}
			//fmt.Printf("Published a message with a message ID: %s\n", id)
		}
	}
	return nil
}

// NewPubSubClient creates a new client connection for pub/sub
func (r *Repository) NewPubSubClient(ctx context.Context, projectID string) (*pubsub.Client, error) {
	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(r.App.GrpcPubSubServer.Conn))
	if err != nil {
		return nil, err
	}
	return client, nil
}

// CreateSubscription creates a subscription for a given client on a topic
func (r *Repository) CreateSubscription(ctx context.Context, subID string, topicName string, c *pubsub.Client) (*pubsub.Subscription, error) {
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

// ReceiveMessage receives the message from pub sub
func (r *Repository) ReceiveMessage(ctx context.Context, sub *pubsub.Subscription) (interface{}, error) {
	var temp []byte
	err := sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		// Do something with message
		temp = m.Data
		m.Ack()
	})
	if err != nil {
		return nil, err
	}

	return temp, nil
}

// CreateTopic receives the message from pub sub
func (r *Repository) CreateTopic(ctx context.Context, topic string, c *pubsub.Client) error {
	_, err := c.CreateTopic(ctx, topic)
	if err != nil {
		return err
	}
	return nil
}

// GenerateRandomMessage for the pub sub
func (r *Repository) GenerateRandomMessage(n uint64, serviceNames *[]string) []Message {
	m := make([]Message, 0)
	for i := uint64(0); i < n; i++ {
		//compose message
		a := Message{
			ServiceName: r.GetRandomServiceName(serviceNames),
			Payload:     r.GetPayLoad(),
			Severity:    r.SeverityToString(r.GetRandomSeverity(0, 4)),
			Timestamp:   time.Now(),
		}
		m = append(m, a)
	}
	return m
}

// SeverityToString converts severity to string
func (r *Repository) SeverityToString(s Severity) string {
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
func (r *Repository) GenerateServicesPool(n uint) *[]string {
	var s []string
	for i := uint(0); i < n; i++ {
		s = append(s, fmt.Sprintf("Service-name:%v", i+1))
	}
	return &s
}

func (r *Repository) InitAndCreateTopic() {

}