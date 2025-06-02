package src

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// Wait for the TCP server to be ready
func waitForTCPServer(address string, retries int) error {
	for i := 0; i < retries; i++ {
		conn, err := net.Dial("tcp", address)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("TCP server not available at %s after %d retries", address, retries)
}

// Test connecting to the running coverage TCP agent
func TestAgentTCPResponse(t *testing.T) {
	err := waitForTCPServer("localhost:8192", 10)
	if err != nil {
		t.Skip("TCP agent not ready: " + err.Error())
	}

	conn, err := net.Dial("tcp", "localhost:8192")
	if err != nil {
		t.Fatalf("Failed to connect to TCP agent: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 8192)
	n, err := conn.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read from agent: %v", err)
	}

	output := string(buf[:n])
	if output == "" {
		t.Skip("No data returned from agent (likely not built with -cover); skipping")
	}

	if strings.Contains(output, "no meta-data available") {
		t.Skip("Binary not built with -cover; skipping TCP agent test")
	}

	if !strings.Contains(output, "SF:") {
		t.Errorf("Unexpected agent response:\n%s", output)
	}
}

// Test logic pipeline: collector → decode → emit
func TestGenerateLcovOutput0(t *testing.T) {
	lcov, err := GenerateLcovOutput()
	if err != nil {
		if strings.Contains(err.Error(), "no meta-data available") {
			t.Skip("Binary not built with -cover; skipping test")
		}
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(lcov, "SF:") {
		t.Errorf("Expected LCOV output to contain 'SF:', got:\n%s", lcov)
	}
}

// Test each step manually: Write → Decode → Emit
func TestFullAgentFlow(t *testing.T) {
	data, err := WriteCoverageToBuffer()
	if err != nil {
		if strings.Contains(err.Error(), "no meta-data available") {
			t.Skip("Binary not built with -cover; skipping full flow test")
		}
		t.Fatalf("Failed to collect coverage: %v", err)
	}

	meta, err := DecodeMeta(data)
	if err != nil {
		t.Fatalf("Failed to decode meta: %v", err)
	}

	counters, err := DecodeCounters(data)
	if err != nil {
		t.Fatalf("Failed to decode counters: %v", err)
	}

	lcov, err := EmitLcov(meta, counters)
	if err != nil {
		t.Fatalf("Failed to emit LCOV: %v", err)
	}

	if !strings.Contains(lcov, "SF:") {
		t.Errorf("Expected LCOV output to contain 'SF:', got:\n%s", lcov)
	}
}
