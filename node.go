package main

import (
	"fmt"
	"net"
	"time"
	"os"
)

const (
	centralNodeIP   = "orion01" // Replace with actual central node IP
	centralNodePort = "10000"     // Replace with actual central node port
)

func main() {
	// Check if nodeID argument is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <nodeID>")
		return
	}

	// Get the nodeID from command-line arguments
	nodeID := os.Args[1]

	sendHeartbeat(nodeID, centralNodeIP, centralNodePort)
}

func sendHeartbeat(nodeID, centralNodeIP string, centralNodePort string) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", centralNodeIP, centralNodePort))
	if err != nil {
		fmt.Printf("Error resolving address: %v\n", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("Error connecting to central node: %v\n", err)
		return
	}
	defer conn.Close()

	for {
		// Heartbeat message
		message := fmt.Sprintf("%s heartbeat", nodeID)

		// Send the heartbeat
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Printf("Error sending heartbeat: %v\n", err)
		} else {
			fmt.Printf("Sent heartbeat: %s\n", message)
		}

		// Wait for 5 seconds before sending the next heartbeat
		time.Sleep(5 * time.Second)
	}
}
