package src

import (
	"log"
	"net"
	"os"
	"os/exec"
)

func init() {
	go startCoverageTCPServer()
}

func startCoverageTCPServer() {
	ln, err := net.Listen("tcp", ":8192")
	if err != nil {
		log.Fatalf("[Agent] Failed to start TCP server: %v", err)
	}
	log.Println("[Agent] Coverage TCP server started on port 8192")

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("[Agent] Connection error: %v", err)
				continue
			}
			go handleCoverageRequest(conn)
		}
	}()
}

func handleCoverageRequest(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("[Agent] Failed to close connection: %v", err)
		}
	}(conn)
	log.Println("[Agent] Received coverage export request")

	// Use Go's built-in covdata tool to convert to LCOV
	cmd := exec.Command("go", "tool", "covdata", "textfmt", "-i=.coverdata", "-o=-")
	cmd.Env = os.Environ() // inherit env, including GOCOVERDIR if needed
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[Agent] Failed to convert coverage: %v", err)
		return
	}

	// Send the LCOV output to the TCP client (e.g., nc)
	_, err = conn.Write(output)
	if err != nil {
		log.Printf("[Agent] Failed to write LCOV to connection: %v", err)
		return
	}

	log.Println("[Agent] LCOV export succeeded")
}
