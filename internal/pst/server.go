package pst

import (
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"context"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"time"
)

// PubSubServer Holds ClientConnection for pstest
type PubSubServer struct {
	Conn *grpc.ClientConn
}

type Client struct {
	*pubsub.Client
}

// Message holds the message structure
type Message struct {
	ServiceName string    `json:"service_name"`
	Payload     string    `json:"payload"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}

var ServiceName string

// StartPubSubFakeServer startup a fake server for pub sub
func StartPubSubFakeServer(port int) (grpcConn *grpc.ClientConn, err error) {
	// Start a fake server running locally at given port.
	srv := pstest.NewServerWithPort(port)
	//defer srv.Close()
	// Connect to the server without using TLS.
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// NewClient creates a new client connection for pubsub
func (s *PubSubServer) NewClientOld(ctx context.Context, projectID string) (*pubsub.Client, error) {
	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(s.Conn))
	if err != nil {
		return nil, err
	}
	return client, nil
}
