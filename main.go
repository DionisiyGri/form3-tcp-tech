package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/form3-tcp-tech/server"
)

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
