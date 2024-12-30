package main

import (
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// func main() {
// 	tcpSocTestingMass()
// }

func tcpSocTestingMass() {
	targetURL := "ws://localhost:43611/api/ws" // Replace with your target WebSocket URL
	var wg sync.WaitGroup

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			connectWebSocket(targetURL, id)
		}(i)
	}

	wg.Wait()
}

func unixSocTesting() {
	socPath := "/tmp/8a7ca4e55c.sock"
	targetURL := "ws://unix/ws"

	conn, err := net.Dial("unix", socPath)
	if err != nil {
		fmt.Println("Error connecting to Unix socket:", err)
		return
	}

	dialer := websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return conn, nil
		},
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		fmt.Println("Error parsing target URL:", err)
		return
	}

	serverConn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket server:", err)
		return
	}

	fmt.Println("Connected to WebSocket server")

	err = serverConn.WriteMessage(websocket.TextMessage, []byte("Hello, WebSocket!"))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

	for {
		messageType, message, err := serverConn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		fmt.Printf("Received message: %s (type: %d)\n", message, messageType)
	}
}

func tcpSocTesting() {
	targetURL := "ws://localhost:43611/api/ws"

	dialer := websocket.DefaultDialer

	u, err := url.Parse(targetURL)
	if err != nil {
		fmt.Println("Error parsing target URL:", err)
		return
	}

	serverConn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket server:", err)
		return
	}

	fmt.Println("Connected to WebSocket server")

	err = serverConn.WriteMessage(websocket.TextMessage, []byte("Hello, WebSocket!"))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

	for {
		messageType, message, err := serverConn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		fmt.Printf("Received message: %s (type: %d)\n", message, messageType)
	}
}

func connectWebSocket(targetURL string, id int) {

	dialer := websocket.DefaultDialer

	u, err := url.Parse(targetURL)
	if err != nil {
		fmt.Printf("Error parsing target URL for connection %d: %v\n", id, err)
		return
	}

	serverConn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Printf("Error connecting to WebSocket server for connection %d: %v\n", id, err)
		return
	}

	err = serverConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Hello, WebSocket from connection %d!", id)))
	if err != nil {
		fmt.Printf("Error sending message for connection %d: %v\n", id, err)
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := serverConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Hello, WebSocket from connection %d at %s!", id, t)))
			if err != nil {
				fmt.Printf("Error sending message for connection %d: %v\n", id, err)
				return
			}
		}
	}
}
