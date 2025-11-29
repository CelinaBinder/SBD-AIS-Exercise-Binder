package main

import (
	"exc8/skeleton/client"
	"exc8/skeleton/server"
	"log"
	"time"
)

func main() {
	go func() {
		// todo start server
		if err := server.StartGrpcServer(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	time.Sleep(1 * time.Second)
	// todo start client
	c, err := client.NewGrpcClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	if err := c.Run(); err != nil {
		log.Fatalf("Client run failed: %v", err)
	}
	println("Orders complete!")
}
