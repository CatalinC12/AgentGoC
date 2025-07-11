package src

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func handleConnection(conn net.Conn) {
	http.Get("http://localhost:8080/flushcov")

	reader := bufio.NewReader(conn)
	firstLine, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		log.Println("[Agent] Failed to read command:", err)
		return
	}
	firstLine = strings.TrimSpace(firstLine)

	switch firstLine {
	case "RESET":
		log.Println("[Agent] Received RESET request")
		resetCoverageData()
		conn.Write([]byte("OK\n"))
		return
	default:
		log.Println("[Agent] Received coverage export request")
	}

	_ = os.Setenv("GOCOVERDIR", ".coverdata")

	tmpOut := ".agent-tmp"
	_ = os.RemoveAll(tmpOut)
	if err := os.MkdirAll(tmpOut, 0755); err != nil {
		log.Println("[Agent] Failed to create temp dir:", err)
		return
	}

	// Merge
	mergeCmd := exec.Command("go", "tool", "covdata", "merge", "-i=.coverdata", "-o="+tmpOut)
	if out, err := mergeCmd.CombinedOutput(); err != nil {
		log.Printf("[Agent] Failed to merge coverage: %v\n%s", err, out)
		return
	}

	// Textfmt
	textFile := filepath.Join(tmpOut, "coverage.out")
	textCmd := exec.Command("go", "tool", "covdata", "textfmt", "-i="+tmpOut, "-o="+textFile)
	if out, err := textCmd.CombinedOutput(); err != nil {
		log.Printf("[Agent] Failed to convert to textfmt: %v\n%s", err, out)
		return
	}

	// Read text and convert to LCOV
	textData, err := os.ReadFile(textFile)
	if err != nil {
		log.Println("[Agent] Failed to read textfmt file:", err)
		return
	}

	lcovData := ConvertTextToLcov(string(textData))

	if _, err := conn.Write([]byte(lcovData)); err != nil {
		log.Println("[Agent] Failed to send LCOV data:", err)
		return
	}

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
					lineNum := coverageParts[0]
					hits := coverageParts[2]
					buf.WriteString("DA:" + string(lineNum) + "," + string(hits) + "\n")
				}
			}
		}
	}
	if currentFile != "" {
		buf.WriteString("end_of_record\n")
	}
	return buf.String()
}
