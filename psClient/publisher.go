package main

import (
	"context"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"fmt"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	ctx := context.Background()
	// Start a fake server running locally at 9001.
	srv := pstest.NewServerWithPort(9001)
	defer srv.Close()
	// Connect to the server without using TLS.
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// Use the connection when creating a pubsub client.
	client, err := pubsub.NewClient(ctx, "project", option.WithGRPCConn(conn))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	//_ = client // TODO: Use the client.
	topic, err := client.CreateTopic(ctx, "lms-topic")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("topic Created")

	defer topic.Stop()
	var results []*pubsub.PublishResult
	r := topic.Publish(ctx, &pubsub.Message{Data: []byte("hello world")})

	results = append(results, r)
	for _, r := range results {
		id, err := r.Get(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Published a message with a message ID: %s\n", id)
	}

	//Client 2
	// Use the connection when creating a pubsub client.
	ctx2 := context.Background()
	client2, err := pubsub.NewClient(ctx2, "project", option.WithGRPCConn(conn))
	if err != nil {
		log.Fatal(err)
	}
	defer client2.Close()

	ok, err := client2.Topic("lms-topic").Exists(ctx2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Topic Exists:", ok)

	t := client.Topic("lms-topic")

	// Create a new subscription to the previously
	//created topic and ensure it never expires.
	_, err = client2.CreateSubscription(ctx2, "lms-sub", pubsub.SubscriptionConfig{
		Topic:            t,
		AckDeadline:      10 * time.Second,
		ExpirationPolicy: time.Duration(0)})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("subscription Created")
	//_ = sub // TODO: Use the subscription

	fmt.Println("trying")

	err = pullMsgs(os.Stdout, "project", "lms-sub", conn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("finsihed")

}

func pullMsgs(w io.Writer, projectID, subID string, conn *grpc.ClientConn) error {
	// projectID := "my-project-id"
	// subID := "my-sub"
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project", option.WithGRPCConn(conn))
	//client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	// Consume 10 messages.
	var mu sync.Mutex
	received := 0
	sub := client.Subscription(subID)
	cctx, cancel := context.WithCancel(ctx)
	fmt.Println("reached")
	err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		defer mu.Unlock()
		fmt.Fprintf(w, "Got message: %q\n", string(msg.Data))
		fmt.Println(string(msg.Data))
		msg.Ack()
		received++
		if received == 10 {
			cancel()
		}
	})
	if err != nil {
		return fmt.Errorf("Receive: %v", err)
	}
	return nil
}
