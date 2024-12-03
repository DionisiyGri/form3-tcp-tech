package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/form3tech-oss/interview-simulator/model"
)

func TestGracefulShutdown(t *testing.T) {
	// Helper function to configure and start the server
	startServer := func(s server, ctx context.Context) {
		go func() {
			if err := s.Start(ctx); err != nil && ctx.Err() == nil {
				t.Fatalf("Failed to start server: %v", err)
			}
		}()
		time.Sleep(time.Second)
	}

	t.Run("ActiveRequestsComplete", func(t *testing.T) {
		s := New(8080)
		ctx, cancel := context.WithCancel(context.Background())

		startServer(s, ctx)

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer conn.Close()

		// Send request
		fmt.Fprintf(conn, "PAYMENT|150\n")
		cancel() // Shutdown

		scanner := bufio.NewScanner(conn)
		if !scanner.Scan() {
			t.Fatal("Failed to read response")
		}
		if resp := scanner.Text(); resp != model.ResponseAccepted {
			t.Errorf("Expected %s response, got %s", model.ResponseAccepted, resp)
		}
	})

	t.Run("StopAcceptingNewConnections", func(t *testing.T) {
		s := New(8080)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		startServer(s, ctx)

		// First connection
		conn1, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
		if err != nil {
			t.Fatalf("Failed to establish connection to server: %v", err)
		}
		defer conn1.Close()

		// Send request
		fmt.Fprintf(conn1, "PAYMENT|150\n")

		// Trigger shutdown
		cancel()
		time.Sleep(100 * time.Millisecond) // Wait to allow server to stop accepting new connections

		// Second connection attempt
		conn2, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
		if err == nil {
			conn2.Close()
			t.Error("Server accepted a new connection after shutdown initiated")
		}
	})

	t.Run("GracePeriodRequestsComplete", func(t *testing.T) {
		s := New(8080)
		ctx, cancel := context.WithCancel(context.Background())

		startServer(s, ctx)

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer conn.Close()

		// Send a request within the grace period
		fmt.Fprintf(conn, "PAYMENT|1500\n")
		cancel() // Trigger shutdown

		// Read the response
		scanner := bufio.NewScanner(conn)
		if !scanner.Scan() {
			t.Fatal("Failed to read response")
		}
		if resp := scanner.Text(); resp != model.ResponseAccepted {
			t.Errorf("Expected %s response, got %s", model.ResponseAccepted, resp)
		}
	})

	t.Run("RejectAfterGracePeriod", func(t *testing.T) {
		s := New(8080)
		ctx, cancel := context.WithCancel(context.Background())

		startServer(s, ctx)

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer conn.Close()

		// Send a request that will exceed the grace period
		cancel()
		fmt.Fprintf(conn, "PAYMENT|5000\n")

		time.Sleep(3500 * time.Millisecond) // Wait for grace period to expire

		// Read the response after grace period
		scanner := bufio.NewScanner(conn)
		if !scanner.Scan() {
			t.Fatal("Failed to read response")
		}
		if resp := scanner.Text(); resp != model.ResponseRejectedCancelled {
			t.Errorf("Expected %s response, got %s", model.ResponseRejectedCancelled, resp)
		}
	})

	t.Run("RequestsNotAcceptedDuringShutdown", func(t *testing.T) {
		s := New(8080)
		ctx, cancel := context.WithCancel(context.Background())

		startServer(s, ctx)
		cancel()

		// Try to establish a connection during shutdown
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
		if err == nil {
			conn.Close()
			t.Error("Server accepted a connection during shutdown")
		}
	})
}
