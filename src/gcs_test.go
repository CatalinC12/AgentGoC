package uleb128

import (
	"bytes"
	"net"
	"strings"
	"testing"
)

type mockConn struct {
	net.Conn
	writeBuf bytes.Buffer
}

func (m *mockConn) Write(p []byte) (int, error) {
	return m.writeBuf.Write(p)
}

func (m *mockConn) Read(p []byte) (int, error) {
	return 0, nil // Not used
}

func (m *mockConn) Close() error {
	return nil
}

func TestGenerateLcovOutput(t *testing.T) {
	output, err := GenerateLcovOutput()
	if err != nil {
		if strings.Contains(err.Error(), "no meta-data available") {
			t.Skip("Binary not built with -cover; skipping test")
		}
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(output, "SF:") {
		t.Error("Expected LCOV output to contain 'SF:'")
	}
}
