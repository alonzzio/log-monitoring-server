package config

type Paragraph struct {
	SentenceCount int
	WordCount     int
}

type ServiceLog struct {
	ServiceNameCharLength uint
}

type PubSub struct {
	ProjectID      string
	TopicID        string
	SubscriptionID string
}
