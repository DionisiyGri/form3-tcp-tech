package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/form3tech-oss/interview-simulator/server"
)

var (
	shutdownSignal = make(chan struct{})
	wg             sync.WaitGroup
	gracePeriod    = 3 * time.Second
)

// TODO:
// 1. Implement a graceful shutdown mechanism, allowing active requests to complete before shutting down.
// 2. The server should stop accepting new connections, but can continue accepting requests.
// 3. The allowed grace period for active requests to complete should be configurable, for example 3 seconds.
// 4. Requests that have been accepted, but not completed after that grace period - should be rejected with: RESPONSE|REJECTED|Cancelled
// 5. Requests that have not been accepted, can be discarded without a response. (ex. slow clients)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdown
		log.Print("Shutdown signal received. Stopping server...")
		close(shutdownSignal)
	}()

	tcpServer := server.New(8080, shutdownSignal)
	if err := tcpServer.Start(); err != nil {
		log.Printf("Server error: %v", err)
	}

	log.Print("exiting...")
}
