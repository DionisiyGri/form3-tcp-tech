package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	shutdownSignal = make(chan struct{})
	wg             sync.WaitGroup
	gracePeriod    = 3 * time.Second
)

// TODO:
// 1.Implement a graceful shutdown mechanism, allowing active requests to complete before shutting down.
// 2. The server should stop accepting new connections, but can continue accepting requests.
// 3. The allowed grace period for active requests to complete should be configurable, for example 3 seconds.
// 4. Requests that have been accepted, but not completed after that grace period - should be rejected with: RESPONSE|REJECTED|Cancelled
// 5. Requests that have not been accepted, can be discarded without a response. (ex. slow clients)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdown
		fmt.Println("\nShutdown signal received. Stopping server...")
		close(shutdownSignal)
	}()

	if err := Start(); err != nil {
		fmt.Println("Server error:", err)
	}

	wg.Wait()
	fmt.Println("All connections completed")
}

func Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8080))
	if err != nil {
		return err
	}
	defer listener.Close()

	//register gracefull shutdown functionality (a kind of)
	stopAccepting := make(chan struct{})
	go func() {
		<-shutdownSignal
		close(stopAccepting)
		listener.Close()
	}()

	fmt.Println("Server is running on port 8080...")

	for {
		conn, err := listener.Accept()
		select {
		case <-stopAccepting:
			return nil
		default:
		}

		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		wg.Add(1)
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	defer wg.Done()

	graceTimer := time.NewTimer(gracePeriod)

	scanner := bufio.NewScanner(conn)
	for {
		select {
		case <-shutdownSignal:
			select {
			case <-graceTimer.C:
				fmt.Println("Grace period expired, rejecting request")
				fmt.Fprintf(conn, "RESPONSE|REJECTED|Cancelled")
				return
			default:
			}
		default:
		}

		if scanner.Scan() {
			request := scanner.Text()
			log.Printf("processing request - %s", request)
			response := handleRequest(request)
			log.Printf("finishing request - %s", request)
			fmt.Fprintf(conn, "%s\n", response)
		} else {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection:", err)
	}
}

func handleRequest(request string) string {
	parts := strings.Split(request, "|")
	if len(parts) != 2 || parts[0] != "PAYMENT" {
		return "RESPONSE|REJECTED|Invalid request"
	}

	amount, err := strconv.Atoi(parts[1])
	if err != nil {
		return "RESPONSE|REJECTED|Invalid amount"
	}

	if amount > 100 {
		processingTime := amount
		if amount > 10000 {
			processingTime = 10000
		}
		time.Sleep(time.Duration(processingTime) * time.Millisecond)
	}
	return "RESPONSE|ACCEPTED|Transaction processed"
}
