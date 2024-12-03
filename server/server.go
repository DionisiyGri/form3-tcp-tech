package server

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/form3tech-oss/interview-simulator/request"
)

type server struct {
	host        string
	port        int
	gracePeriod time.Duration
	wg          *sync.WaitGroup
}

func New() server {
	return server{
		host:        "localhost",
		port:        8080,
		gracePeriod: 2 * time.Second,
		wg:          new(sync.WaitGroup),
	}
}

// Start tcp server and ready to accept connections
func (s server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		log.Print("Shutting down server: closing listener")
		listener.Close()
	}()

	log.Printf("Server is running on port %d...", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(ctx, conn)
		}()
	}
	s.wg.Wait()
	log.Println("All connections closed. Server shutdown complete.")
	return nil
}

func (s server) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		log.Printf("closing connection from %s", conn.RemoteAddr())
		if err := conn.Close(); err != nil {
			log.Printf("cannot close connection. err = %v", err)
		}
	}()

	gracePeriod := time.After(s.gracePeriod)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			select {
			case <-gracePeriod:
				log.Printf("Grace period expired, rejecting request from %s", conn.RemoteAddr())
				fmt.Fprintf(conn, "%s\n", "RESPONSE|REJECTED|Cancelled")
				return
			default:
			}
		default:
		}

		req := scanner.Text()
		log.Printf("[START] Connection address: %s, request: %s", conn.RemoteAddr(), req)

		response := request.Handle(req)
		log.Printf("[FINISH]Connection address: %s, response: %s", conn.RemoteAddr(), response)

		fmt.Fprintf(conn, "%s\n", response)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from connection: %v", err)
	}
}
