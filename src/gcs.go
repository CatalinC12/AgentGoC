package src

import (
	_ "bytes"
	"fmt"
	_ "io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	_ "time"
)

func StartCoverageAgent() {
	ln, err := net.Listen("tcp", ":8192")
	if err != nil {
		log.Fatalf("[Agent] Failed to start TCP server: %v", err)
	}
	log.Println("[Agent] Coverage TCP server started on port 8192")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[Agent] Failed to accept connection: %v", err)
			continue
		}
		go handleCoverageRequest(conn)
	}
}

func handleCoverageRequest(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("[Agent] Failed to close connection: %v", err)
		}
	}(conn)
	log.Println("[Agent] Received coverage request")

	err := flushCoverageData(".coverdata")
	if err != nil {
		log.Printf("[Agent] Error flushing coverage data: %v", err)
		return
	}

	lcov, err := convertCoverageToLCOV(".coverdata")
	if err != nil {
		log.Printf("[Agent] Error converting to LCOV: %v", err)
		return
	}

	_, err = conn.Write([]byte(lcov))
	if err != nil {
		log.Printf("[Agent] Failed to write LCOV to connection: %v", err)
	}
}

func flushCoverageData(dir string) error {
	err := os.Setenv("GOCOVERDIR", dir)
	if err != nil {
		return err
	}

	// call runtime coverage flush helpers
	if err := exec.Command("go", "tool", "covdata", "func", "-i", dir).Run(); err != nil {
		return fmt.Errorf("failed to trigger coverage tool: %w", err)
	}
	return nil
}

func convertCoverageToLCOV(dir string) (string, error) {
	coverageFile := filepath.Join(dir, "coverage.out")
	cmd := exec.Command("go", "tool", "covdata", "textfmt", "-i", dir, "-o", coverageFile)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run go tool covdata: %w", err)
	}

	data, err := os.ReadFile(coverageFile)
	if err != nil {
		return "", fmt.Errorf("failed to read coverage.out: %w", err)
	}

	return parseCoverProfileToLCOV(string(data)), nil
}

func parseCoverProfileToLCOV(input string) string {
	var buf strings.Builder
	lines := strings.Split(input, "\n")
	currentFile := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "mode:") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		file, rest := parts[0], parts[1]
		if file != currentFile {
			if currentFile != "" {
				buf.WriteString("end_of_record\n")
			}
			currentFile = file
			buf.WriteString(fmt.Sprintf("SF:%s\n", file))
		}
		fields := strings.Fields(rest)
		if len(fields) < 2 {
			continue
		}
		pos := strings.Split(fields[0], ",")
		start := strings.Split(pos[0], ".")[0]
		count := fields[1]
		buf.WriteString(fmt.Sprintf("DA:%s,%s\n", start, count))
	}
	buf.WriteString("end_of_record\n")
	return buf.String()
}

func init() {
	go StartCoverageAgent()
}
