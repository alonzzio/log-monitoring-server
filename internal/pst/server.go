package pst

import (
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/grpc"
)

// PubSubServer Holds ClientConnection for pstest
type PubSubServer struct {
	Conn *grpc.ClientConn
}

// StartPubSubFakeServer startup a fake server for pub sub
func StartPubSubFakeServer(port int) (*grpc.ClientConn, *pstest.Server, error) {
	// Start a fake server running locally at given port.
	srv := pstest.NewServerWithPort(port)
	//defer srv.Close()
	// Connect to the server without using TLS.
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}

	return conn, srv, nil
}
