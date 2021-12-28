package main

import (
	"cloud.google.com/go/pubsub/pstest"
	"log"
	"time"
)

func main() {
	log.Println("Pub Sub Fake server Starting at Port:9001")
	// Start a fake server running locally at 9001.
	srv := pstest.NewServerWithPort(9001)
	defer srv.Close()

	time.Sleep(100 * time.Second)
	log.Println("Fake Pub Sub Server Shutting Down.")
}
