package src

import (
	"log"
	"net"
	"os"
)

// Automatically start the TCP server when the agent is imported
func init() {
	go startTCPServer()
}

func startTCPServer() {
	const port = "8192"
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Agent failed to start TCP server: %v", err)
	}
	log.Printf("[Agent] Coverage TCP server started on port %s", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting TCP connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	log.Println("[Agent] Received coverage request")

	// Grab the latest in-memory coverage buffers
	metaBuf, counterBuf, err := WriteCoverageToBuffers()
	if err != nil {
		log.Printf("[Agent] Coverage buffer error: %v", err)
		return
	}

	// Decode metadata & counters
	meta, err := DecodeFullMetaFile(metaBuf)
	if err != nil {
		log.Printf("[Agent] DecodeMeta error: %v", err)
		return
	}
	counters, err := DecodeCounters(counterBuf)
	if err != nil {
		log.Printf("[Agent] DecodeCounters error: %v", err)
		return
	}

	// Emit LCOV and send over TCP
	lcov, err := EmitLcov(meta, counters)
	if err != nil {
		log.Printf("[Agent] EmitLcov error: %v", err)
		return
	}

	// save a copy locally
	if err := os.WriteFile("coverage.lcov", []byte(lcov), 0644); err != nil {
		log.Printf("[Agent] Failed to write coverage.lcov: %v", err)
	}

	if _, err := conn.Write([]byte(lcov)); err != nil {
		log.Printf("[Agent] Error sending LCOV data: %v", err)
	}
}
