package server

import (
	"bufio"
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
	shutdownSig chan struct{}
}

func New(port int, shutdownSig chan struct{}) server {
	return server{
		host:        "localhost",
		port:        port,
		gracePeriod: 3 * time.Second,
		wg:          new(sync.WaitGroup),
		shutdownSig: shutdownSig,
	}
}

// Start tcp server and ready to accept connections
func (s server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return err
	}

	//register gracefull shutdown functionality (a kind of)
	stopAccepting := make(chan struct{})
	go func() {
		<-s.shutdownSig
		close(stopAccepting)
		listener.Close()
	}()

	log.Printf("Server is running on port %d...", s.port)

	for {
		conn, err := listener.Accept()
		select {
		case <-stopAccepting:
			log.Print("Stop receiving connections")
			return nil
		default:
		}

		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(conn)
		}()
		s.wg.Wait()
	}
}

func (s server) handleConnection(conn net.Conn) {
	defer func() {
		log.Print("closing connection")
		if err := conn.Close(); err != nil {
			log.Printf("cannot close connection. err = %v", err)
		}
	}()

	graceTimer := time.NewTimer(s.gracePeriod)
	defer graceTimer.Stop()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		select {
		case <-s.shutdownSig:
			select {
			case <-graceTimer.C:
				log.Print("Grace period expired, rejecting request")
				fmt.Fprintf(conn, "%s\n", "RESPONSE|REJECTED|Cancelled")
				return
			default:
				//processing request during grace period
			}
		default:
		}

		req := scanner.Text()
		log.Printf("request start: %s", req)

		response := request.Handle(req)
		log.Printf("request finish:%s", req)

		fmt.Fprintf(conn, "%s\n", response)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from connection: %v", err)
	}
}
