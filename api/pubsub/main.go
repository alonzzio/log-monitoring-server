package main

import (
	"context"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"fmt"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"log"
)

func main() {
	ctx := context.Background()
	// Start a fake server running locally.
	srv := pstest.NewServer()
	defer srv.Close()
	// Connect to the server without using TLS.
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		// TODO: Handle error.
		log.Fatalln(err)
	}
	defer conn.Close()
	// Use the connection when creating a pubsub client.
	client, err := pubsub.NewClient(ctx, "project", option.WithGRPCConn(conn))
	if err != nil {
		// TODO: Handle error.
		log.Fatalln(err)
	}
	defer client.Close()
	c := client // TODO: Use the client.
	fmt.Println(c)
	// todo functionalities
}
