package src

import (
	"log"
	"net"
	"os"
	"runtime/coverage"
)

// Automatically start the TCP server when the agent is imported
func init() {
	go startTCPServer()
}

func startTCPServer() {
	ln, err := net.Listen("tcp", ":8192")
	if err != nil {
		log.Fatalf("Agent failed to start TCP server: %v", err)
	}
	log.Println("[Agent] Coverage TCP server started on port 8192")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[Agent] Error accepting TCP connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("[Agent] Error closing TCP connection: %v", err)
		}
	}(conn)
	log.Println("[Agent] Received coverage request")

	dir := os.Getenv("GOCOVERDIR")
	if dir == "" {
		dir = ".coverdata"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[Agent] mkdir %s: %v", dir, err)
	}
	if err := coverage.WriteMetaDir(dir); err != nil {
		log.Printf("[Agent] WriteMetaDir error: %v", err)
	}
	if err := coverage.WriteCountersDir(dir); err != nil {
		log.Printf("[Agent] WriteCountersDir error: %v", err)
	} else {
		log.Printf("[Agent] Flushed .covmeta and .covcounters into %s", dir)
	}

	metaBuf, counterBuf, err := WriteCoverageToBuffers()
	if err != nil {
		log.Printf("[Agent] Coverage buffer error: %v", err)
		return
	}
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

	// Emit LCOV
	lcov, err := EmitLcov(meta, counters)
	if err != nil {
		log.Printf("[Agent] EmitLcov error: %v", err)
		return
	}

	// Save a human-readable copy
	if err := os.WriteFile("coverage.lcov", []byte(lcov), 0644); err != nil {
		log.Printf("[Agent] Failed to write coverage.lcov: %v", err)
	}

	// Stream it back
	if _, err := conn.Write([]byte(lcov)); err != nil {
		log.Printf("[Agent] Error sending LCOV data: %v", err)
	}
}
