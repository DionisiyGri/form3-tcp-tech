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
	s := New()
	s.gracePeriod = 3 * time.Second

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := s.Start(ctx); err != nil && ctx.Err() == nil {
			t.Fatalf("Failed to start server: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond) // give some time for server to start

	t.Run("ActiveRequestsComplete", func(t *testing.T) {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port)) //create connection
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer conn.Close()

		fmt.Fprintf(conn, "PAYMENT|150\n") //sending request
		cancel()                           //shutdown

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			resp := scanner.Text()
			if resp != model.ResponseAccepted {
				t.Errorf("Expected %s response, got %s", model.ResponseAccepted, resp)
			}
		}
		if scanner.Err() != nil {
			t.Errorf("Failed to scan response: %v", err)
		}
	})
	t.Run("StopAcceptingNewConnections", func(t *testing.T) {
		conn1, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port)) // creating 1st connection
		if err != nil {
			t.Fatalf("Failed to establish connection to server: %v", err)
		}
		defer conn1.Close()

		fmt.Fprintf(conn1, "PAYMENT|150\n") //sending request

		cancel() //shutdown
		time.Sleep(100 * time.Millisecond)

		conn2, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port)) // creating 2nd connection
		if err == nil {                                                     // Connection shouldnt be accepted
			conn2.Close()
			t.Error("Server accepted a new connection after shutdown initiated")
		}
	})

	t.Run("GracePeriodRequestsComplete", func(t *testing.T) {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer conn.Close()

		fmt.Fprintf(conn, "PAYMENT|1500\n") // 1.5 seconds
		cancel()                            // Trigger shutdown

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			resp := scanner.Text()
			if resp != model.ResponseAccepted {
				t.Errorf("Expected %s response, got %s", model.ResponseAccepted, resp)
			}
		}
		if scanner.Err() != nil {
			t.Errorf("Failed to scan response: %v", err)
		}
	})

	t.Run("RejectAfterGracePeriod", func(t *testing.T) {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer conn.Close()

		fmt.Fprintf(conn, "PAYMENT|5000\n") // 5sec request  (> 3 seconds)
		cancel()

		time.Sleep(4 * time.Second) // Wait for grace period to expire

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			resp := scanner.Text()
			if resp != model.ResponseRejectedCancelled {
				t.Errorf("Expected %s response, got %s", model.ResponseRejectedCancelled, resp)
			}
		}
		if scanner.Err() != nil {
			t.Errorf("Failed to scan response: %v", err)
		}
	})

	t.Run("RequestsNotAcceptedDuringShutdown", func(t *testing.T) {
		cancel() // Trigger shutdown immediately

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
		if err == nil { // Connection should not be established during shutdown
			conn.Close()
			t.Error("Server accepted a connection during shutdown")
		}
	})

}
