package main

import (
	"fmt"
	"net"
	"sync"
	"time"
	"os/exec"
)

const (
	listenPort       = "10000"  // Port for listening to heartbeats
	heartbeatTimeout = 15 * time.Second // Time to wait before considering a node as failed
)

var (
	nodes       = make(map[string]time.Time) // Store the last heartbeat time for each node
	addr        = make(map[string]string)
	nodesMutex  sync.Mutex                    // Mutex to protect access to the nodes map
)

func main() {
	// Start the monitoring routine in the background
	go monitorNodes()

	// Start listening for heartbeat messages
	listenForHeartbeats(listenPort)
}

// Listen for heartbeat messages from the nodes
func listenForHeartbeats(port string) {
	address, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		fmt.Printf("Error resolving address: %v\n", err)
		return
	}

	conn, err := net.ListenUDP("udp", address)
	if err != nil {
		fmt.Printf("Error listening on port %s: %v\n", port, err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Error receiving message: %v\n", err)
			continue
		}

		heartbeatMessage := string(buffer[:n])
		fmt.Printf("Received heartbeat from %s: %s\n", remoteAddr, heartbeatMessage)

		nodeID := parseNodeID(heartbeatMessage)

		// Update the node's last heartbeat time
		updateNodeHeartbeat(nodeID)
		addr[nodeID] = remoteAddr.IP.String()
	}
}

// Parse the node ID from the heartbeat message
func parseNodeID(message string) string {
	// In this example, the nodeID is the first part of the message
	// e.g., "node_1 heartbeat" -> nodeID = "node_1"
	var nodeID string
	fmt.Sscanf(message, "%s", &nodeID)
	return nodeID
}

// Update the last heartbeat time for a given node
func updateNodeHeartbeat(nodeID string) {
	nodesMutex.Lock()
	defer nodesMutex.Unlock()

	// Update the node's last heartbeat time to the current time
	nodes[nodeID] = time.Now()
}

// Monitor the nodes and restart any that have missed a heartbeat
func monitorNodes() {
	for {
		time.Sleep(1 * time.Second) // Check every second

		currentTime := time.Now()

		nodesMutex.Lock()
		for nodeID, lastHeartbeat := range nodes {
			if currentTime.Sub(lastHeartbeat) > heartbeatTimeout {
				fmt.Printf("%s missed heartbeats, restarting...\n", nodeID)
				restartNode(nodeID)
				// Remove the node from the map after restarting
				delete(nodes, nodeID)
			}
		}
		nodesMutex.Unlock()
	}
}

// Restart the node via SSH
func restartNode(nodeID string) {
	// Get the node's IP address
	nodeIP, exists := addr[nodeID]
	if !exists {
		fmt.Printf("No IP address found for %s\n", nodeID)
		return
	}

	// Define the SSH command to restart the node (replace with actual username and path)
	user := "jcai28"          // Replace with your SSH username
	nodeExecutablePath := "./node"   // Replace with the actual path to the node executable on the remote server

	// SSH command: ssh user@nodeIP "cd /path/to/node && ./node"
	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", user, nodeIP), fmt.Sprintf("cd L3-Template && %s", nodeExecutablePath))

	// Run the command and capture output and errors
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to restart node %s (%s): %v, output: %s\n", nodeID, nodeIP, err, string(output))
		return
	}

	fmt.Printf("Node %s restarted successfully at %s. Output: %s\n", nodeID, nodeIP, string(output))
}
