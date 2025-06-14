package src

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
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
			log.Printf("Error closing TCP connection: %v", err)
		}
	}(conn)
	log.Println("[Agent] Received coverage request")

	// Grab in-memory coverage buffers
	metaBuf, counterBuf, err := WriteCoverageToBuffers()
	if err != nil {
		log.Printf("[Agent] Coverage buffer error: %v", err)
		return
	}

	// Flush them as files (so you see covmeta & covcounters on disk)
	if err := flushToDisk(metaBuf, counterBuf); err != nil {
		log.Printf("[Agent] Error writing coverage files: %v", err)
		// but continue to decode and emit LCOV anyway
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

	// Emit LCOV
	lcov, err := EmitLcov(meta, counters)
	if err != nil {
		log.Printf("[Agent] EmitLcov error: %v", err)
		return
	}

	// Save a copy locally
	if err := os.WriteFile("coverage.lcov", []byte(lcov), 0644); err != nil {
		log.Printf("[Agent] Failed to write coverage.lcov: %v", err)
	}

	// Send it back over TCP
	if _, err := conn.Write([]byte(lcov)); err != nil {
		log.Printf("[Agent] Error sending LCOV data: %v", err)
	}
}

// flushToDisk writes the raw meta and counter buffers into GOCOVERDIR (or .coverdata).
func flushToDisk(metaBuf, counterBuf []byte) error {
	dir := os.Getenv("GOCOVERDIR")
	if dir == "" {
		dir = ".coverdata"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	metaPath := filepath.Join(dir, "covmeta")
	if err := os.WriteFile(metaPath, metaBuf, 0644); err != nil {
		return fmt.Errorf("write %s: %w", metaPath, err)
	}

	counterPath := filepath.Join(dir, "covcounters")
	if err := os.WriteFile(counterPath, counterBuf, 0644); err != nil {
		return fmt.Errorf("write %s: %w", counterPath, err)
	}

	log.Printf("[Agent] Wrote %s and %s", metaPath, counterPath)
	return nil
}
