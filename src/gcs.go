package uleb128

import (
	"fmt"
	"log"
	"net"
	_ "os"
)

// Automatically start the TCP server when the agent is imported
func init() {
	go startTCPServer()
}

// Starts the TCP server on port 8192 and listens for coverage requests
func startTCPServer() {
	port := "8192"
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Agent failed to start TCP server: %v", err)
	}
	fmt.Printf("[Agent] Coverage TCP server started on port %s\n", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting TCP connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

// GenerateLcovOutput Handles a TCP request from WuppieFuzz
func GenerateLcovOutput() (string, error) {
	data, err := WriteCoverageToBuffer()
	if err != nil {
		return "", err
	}

	meta, err := DecodeMeta(data)
	if err != nil {
		return "", err
	}

	counters, err := DecodeCounters(data)
	if err != nil {
		return "", err
	}

	return EmitLcov(meta, counters)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("[Agent] Received coverage request")

	lcov, err := GenerateLcovOutput()
	if err != nil {
		log.Printf("Error generating LCOV output: %v", err)
		return
	}

	_, err = conn.Write([]byte(lcov))
	if err != nil {
		log.Printf("Error sending LCOV data: %v", err)
	}
}
