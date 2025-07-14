package src

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	go startCoverageAgent()
}

func startCoverageAgent() {
	go func() {
		ln, err := net.Listen("tcp", ":8192")
		if err != nil {
			log.Println("[Agent] Failed to start TCP listener:", err)
			return
		}
		log.Println("[Agent] Coverage TCP server started on port 8192")

		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println("[Agent] Accept error:", err)
				continue
			}
			go handleConnection(conn)
		}
	}()
}

func sendBlock(conn net.Conn, blockType byte, data []byte) error {
	log.Printf("[Agent] Sent block type: 0x%X (%d bytes)", blockType, len(data))
	length := uint32(len(data))
	header := []byte{
		blockType,
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}
	if _, err := conn.Write(header); err != nil {
		return err
	}
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

func isTimeout(err error) bool {
	var ne net.Error
	ok := errors.As(err, &ne)
	return ok && ne.Timeout()
}

func handleConnection(conn net.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println("[Agent] Failed to close connection:", err)
		} else {
			log.Println("[Agent] Connection closed")
		}
	}()

	log.Println("[Agent] Connection accepted")

	http.Get("http://localhost:8080/flushcov")

	reader := bufio.NewReader(conn)
	_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	firstLine, err := reader.ReadString('\n')
	if err != nil && err != io.EOF && !isTimeout(err) {
		log.Println("[Agent] Failed to read command:", err)
		return
	}
	firstLine = strings.TrimSpace(firstLine)

	if firstLine == "RESET" {
		log.Println("[Agent] Received RESET request")
		resetCoverageData()
		conn.Write([]byte("OK\n"))
		return
	}

	log.Println("[Agent] Received coverage export request")

	_ = os.Setenv("GOCOVERDIR", ".coverdata")

	tmpOut := ".agent-tmp"
	_ = os.RemoveAll(tmpOut)
	if err := os.MkdirAll(tmpOut, 0755); err != nil {
		log.Println("[Agent] Failed to create temp dir:", err)
		return
	}

	// Merge coverage
	mergeCmd := exec.Command("go", "tool", "covdata", "merge", "-i=.coverdata", "-o="+tmpOut)
	if out, err := mergeCmd.CombinedOutput(); err != nil {
		log.Printf("[Agent] Failed to merge coverage: %v\n%s", err, out)
		return
	}

	// Convert to text
	textFile := filepath.Join(tmpOut, "coverage.out")
	textCmd := exec.Command("go", "tool", "covdata", "textfmt", "-i="+tmpOut, "-o="+textFile)
	if out, err := textCmd.CombinedOutput(); err != nil {
		log.Printf("[Agent] Failed to convert to textfmt: %v\n%s", err, out)
		return
	}

	// Read and send LCOV in blocks
	textData, err := os.ReadFile(textFile)
	if err != nil {
		log.Println("[Agent] Failed to read textfmt file:", err)
		return
	}

	lcovData := ConvertTextToLcov(string(textData))
	sendBlock(conn, 0x01, []byte("LCOV coverage data"))
	sendBlock(conn, 0x02, []byte(lcovData))
	sendBlock(conn, 0xFF, nil)

	log.Println("[Agent] LCOV export succeeded")
}

func resetCoverageData() {
	dir := ".coverdata"

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("[Agent] Failed to read %s: %v", dir, err)
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "covcounters") {
			path := filepath.Join(dir, name)
			err := os.RemoveAll(path)
			if err != nil {
				log.Printf("[Agent] Failed to delete %s: %v", path, err)
			} else {
				log.Printf("[Agent] Deleted %s", path)
			}
		}
	}

	err = os.RemoveAll(".agent-tmp")
	if err != nil {
		log.Printf("[Agent] Failed to delete .agent-tmp: %v", err)
	}

	log.Println("[Agent] Coverage data reset complete")
}

func ConvertTextToLcov(input string) string {
	var buf bytes.Buffer
	var currentFile string
	for _, line := range bytes.Split([]byte(input), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		if bytes.HasPrefix(line, []byte("mode:")) {
			continue
		}
		parts := bytes.Split(line, []byte(":"))
		if len(parts) < 2 {
			continue
		}
		fileFunc := string(parts[0])
		rest := string(parts[1])
		if bytes.HasPrefix(line, []byte("SF:")) || filepath.Ext(fileFunc) == ".go" {
			if fileFunc != currentFile {
				if currentFile != "" {
					buf.WriteString("end_of_record\n")
				}
				currentFile = fileFunc
				buf.WriteString("SF:" + fileFunc + "\n")
			}
			if len(parts) >= 2 {
				coverageParts := bytes.Fields([]byte(rest))
				if len(coverageParts) >= 3 {
					lineCol := strings.Split(string(coverageParts[0]), ".")
					lineNum := lineCol[0]
					hits := string(coverageParts[2])
					buf.WriteString("DA:" + lineNum + "," + hits + "\n")
				}
			}
		}
	}
	if currentFile != "" {
		buf.WriteString("end_of_record\n")
	}
	return buf.String()
}
