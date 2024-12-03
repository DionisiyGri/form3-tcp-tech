package server

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/form3tech-oss/interview-simulator/model"
	"github.com/form3tech-oss/interview-simulator/request"
)

type server struct {
	host        string
	port        int
	gracePeriod time.Duration
	wg          *sync.WaitGroup
}

// New create a new server instance
func New() server {
	return server{
		host:        "localhost",
		port:        8080,
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
			select {
			case <-gracePeriod.C:
				log.Printf("Grace period expired, rejecting further requests from %s", conn.RemoteAddr())
				fmt.Fprintf(conn, "%s\n", model.ResponseRejectedCancelled)
				return
			default:
				// During the grace period, allow active requests to finish
				if scanner.Scan() {
					req := scanner.Text()
					log.Printf("[grace period request]: %s, Address: %s", req, conn.RemoteAddr())

					// Handle request, but enforce timeout behavior
					response := make(chan string, 1)
					go func() {
						response <- request.Handle(req)
					}()

					select {
					case resp := <-response:
						log.Printf("[grace period response]: %s, Address: %s", resp, conn.RemoteAddr())
						fmt.Fprintf(conn, "%s\n", resp)
						return
					case <-gracePeriod.C:
						log.Printf("Grace period expired during processing, rejecting request from %s", conn.RemoteAddr())
						fmt.Fprintf(conn, "%s\n", model.ResponseRejectedCancelled)
						return
					}
				} else {
					log.Printf("No further requests to handle from %s", conn.RemoteAddr())
					return
				}
			}
		default:
			// Handle normal request processing
			if scanner.Scan() {
				req := scanner.Text()
				log.Printf("[request]: %s, Address: %s", req, conn.RemoteAddr())

				response := request.Handle(req)
				log.Printf("[response]: %s, Address: %s", response, conn.RemoteAddr())

				fmt.Fprintf(conn, "%s\n", response)
				return
			} else {
				log.Printf("Connection closed or no more requests from %s", conn.RemoteAddr())
				return
			}
		}
	}
}
