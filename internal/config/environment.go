package config

type Paragraph struct {
	SentenceCount int
	WordCount     int
}

type ServiceLog struct {
	ServiceNameCharLength uint
}

type PubSub struct {
	ProjectID         string
	TopicID           string
	SubscriptionID    string
	ServicePublishers uint
	ServiceNamePool   uint
	MessageBatch      uint
	MessageFrequency  uint // for time.Duration in milliseconds
}

type DataAccessLayer struct {
	PortNumber string
}

type DataCollectionLayer struct {
	Workers         int
	JobsBuffer      int
	ResultBuffer    int
	ReceiverTimeOut int
}
