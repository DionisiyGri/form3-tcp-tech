package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/form3-tcp-tech/server"
)

// TODO:
// 1. Implement a graceful shutdown mechanism, allowing active requests to complete before shutting down. +
// 2. The server should stop accepting new connections, but can continue accepting requests. +
// 3. The allowed grace period for active requests to complete should be configurable, for example 3 seconds. +
// 4. Requests that have been accepted, but not completed after that grace period - should be rejected with: RESPONSE|REJECTED|Cancelled +
// 5. Requests that have not been accepted, can be discarded without a response. (ex. slow clients) +

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-shutdown
		log.Print("Shutdown signal received. Stopping server...")
		cancel()
	}()

	tcpServer := server.New(8080)
	if err := tcpServer.Start(ctx); err != nil {
		log.Printf("Server error: %v", err)
	}

	log.Print("exiting...")
}
