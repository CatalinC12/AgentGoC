package src

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
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

// GenerateLcovOutput handles a TCP request from WuppieFuzz
func GenerateLcovOutput() (string, error) {
	const maxAttempts = 3
	const delay = 200 * time.Millisecond

	for i := 0; i < maxAttempts; i++ {
		metaBuf, counterBuf, err := WriteCoverageToBuffers()
		if err != nil {
			log.Printf("[Agent] Coverage buffer error: %v", err)
			time.Sleep(delay)
			continue
		}

		meta, err := DecodeFullMetaFile(metaBuf)

		if err != nil {
			log.Printf("[Agent] DecodeMeta error (attempt %d): %v", i+1, err)
			time.Sleep(delay)
			continue
		}

		counters, err := DecodeCounters(counterBuf)
		if err != nil {
			log.Printf("[Agent] DecodeCounters error (attempt %d): %v", i+1, err)
			time.Sleep(delay)
			continue
		}

		return EmitLcov(meta, counters)
	}

	return "", fmt.Errorf("failed to generate LCOV output after retries")
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}(conn)
	fmt.Println("[Agent] Received coverage request")

	lcov, err := GenerateLcovOutput()
	if err != nil {
		log.Printf("Error generating LCOV output: %v", err)
		return
	}

	err = os.WriteFile("report.lcov", []byte(lcov), 0644)
	if err != nil {
		log.Printf("Error writing LCOV file: %v", err)
	}

	_, err = conn.Write([]byte(lcov))
	if err != nil {
		log.Printf("Error sending LCOV data: %v", err)
	}
}
