package pst

import (
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/grpc"
)

// PubsubFakeServer startup a fake server for pub sub
func PubsubFakeServer() (grpcConn *grpc.ClientConn, err error) {
	// Start a fake server running locally at 9001.
	srv := pstest.NewServerWithPort(9001)
	//defer srv.Close()
	// Connect to the server without using TLS.
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return conn, nil
}
