// just a simple client for local calls and fast tests
package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	fmt.Fprintf(conn, "PAYMENT|2000\n")

	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error read response", err)
		return
	}
	fmt.Print("Server response: ", message)
}
