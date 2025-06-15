package src

import (
	"log"
	"net"
	"os"
)

func init() {
	go startTCPServer()
}

func startTCPServer() {
	ln, err := net.Listen("tcp", ":8192")
	if err != nil {
		log.Fatalf("[Agent] failed to start TCP server: %v", err)
	}
	log.Println("[Agent] Coverage TCP server started on port 8192")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[Agent] accept error: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("[Agent] failed to close connection: %v", err)
		}
	}(conn)
	log.Println("[Agent] Received coverage export request")

	metaBuf, counterBuf, err := WriteCoverageToBuffers()
	if err != nil {
		log.Printf("[Agent] WriteCoverageToBuffers error: %v", err)
		return
	}
	meta, err := DecodeMeta(metaBuf)
	if err != nil {
		log.Printf("[Agent] DecodeMeta error: %v", err)
		return
	}
	counters, err := DecodeCounters(counterBuf)
	if err != nil {
		log.Printf("[Agent] DecodeCounters error: %v", err)
		return
	}
	lcov, err := EmitLcov(meta, counters)
	if err != nil {
		log.Printf("[Agent] EmitLcov error: %v", err)
		return
	}

	_ = os.WriteFile("coverage.lcov", []byte(lcov), 0644)
	_, _ = conn.Write([]byte(lcov))
	log.Println("[Agent] LCOV report sent")
}
