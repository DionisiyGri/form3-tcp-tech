package server

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/form3-tcp-tech/model"
	"github.com/form3-tcp-tech/request"
)

type server struct {
	host        string
	port        int
	gracePeriod time.Duration
	wg          *sync.WaitGroup
}

// New create a new server instance
func New(port int) server {
	return server{
		host:        "localhost",
		port:        port,
		gracePeriod: 2 * time.Second,
		wg:          new(sync.WaitGroup),
	}
}

// Start tcp server and ready to accept connections
// Listens to context cancelation to shoutdown and cleanup
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
				log.Print("stop accepting new connections")
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
		log.Printf("Closing connection with %s", conn.RemoteAddr())
		if err := conn.Close(); err != nil {
			log.Printf("Close connection error: %v", err)
		}
	}()

	gracePeriod := time.NewTimer(s.gracePeriod)
	defer gracePeriod.Stop()

	scanner := bufio.NewScanner(conn)

	for {
		select {
		case <-ctx.Done():
			if err := s.handleGracePeriodExpired(conn, gracePeriod, scanner); err != nil {
				log.Printf("Error during processing grace period requests: %v", err)
				return
			}
			return
		default:
			if err := s.handleNormalRequest(scanner, conn); err != nil {
				log.Printf("Error handling request: %v", err)
				return
			}
			return
		}
	}
}

func (s server) handleGracePeriodExpired(conn net.Conn, gracePeriod *time.Timer, scanner *bufio.Scanner) error {
	select {
	case <-gracePeriod.C:
		log.Printf("Grace period expired, rejecting further requests from %s", conn.RemoteAddr())
		fmt.Fprintf(conn, "%s\n", model.ResponseRejectedCancelled)
		return fmt.Errorf("grace period expired")
	default:
		// Allow active requests to finish during the grace period
		if scanner.Scan() {
			req := scanner.Text()
			log.Printf("[grace period request]: %s, Address: %s", req, conn.RemoteAddr())

			// enforce timeout behavior
			response := make(chan string, 1)
			go func() {
				response <- request.Handle(req)
			}()

			select {
			case resp := <-response:
				log.Printf("[grace period response]: %s, Address: %s", resp, conn.RemoteAddr())
				fmt.Fprintf(conn, "%s\n", resp)
				return nil
			case <-gracePeriod.C:
				log.Printf("Grace period expired during processing, rejecting request from %s", conn.RemoteAddr())
				fmt.Fprintf(conn, "%s\n", model.ResponseRejectedCancelled)
				return fmt.Errorf("grace period expired during processing")
			}
		} else {
			log.Printf("No further requests to handle from %s", conn.RemoteAddr())
			return nil
		}
	}
}

// handleNormalRequest processes a normal request outside of the grace period.
func (s server) handleNormalRequest(scanner *bufio.Scanner, conn net.Conn) error {
	if scanner.Scan() {
		req := scanner.Text()
		log.Printf("[request]: %s, Address: %s", req, conn.RemoteAddr())

		response := request.Handle(req)
		log.Printf("[response]: %s, Address: %s", response, conn.RemoteAddr())
		fmt.Fprintf(conn, "%s\n", response)
		return nil
	} else {
		return fmt.Errorf("connection closed or no more requests")
	}
}
