// just a simple client for local calls and smoke tests
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func send(i int, conn net.Conn) {
	message := "PAYMENT|3000"
	log.Printf("Request: %d, message: %s, RemoteAddress: %s\n", i, message, conn.LocalAddr())
	fmt.Fprintf(conn, message+"\n")

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		//log.Printf("Request %d: Error reading response: %v\n", i, err)
		return
	}
	log.Printf("Request: %d, response: %s, RemoteAddress: %s\n", i, response, conn.LocalAddr())
}

func main() {
	var wg sync.WaitGroup

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()
	conn2, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn2.Close()

	for i := 1; i <= 6; i++ {
		// if i == 4 || i == 5 {
		// 	time.Sleep(1000 * time.Millisecond)
		// }
		time.Sleep(500 * time.Millisecond)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if i == 4 || i == 5 {
				send(i, conn2)
			}
			send(i, conn)
		}()
	}
	wg.Wait()

	log.Print("sleeping before exiting")
	time.Sleep(20 * time.Second)

}
