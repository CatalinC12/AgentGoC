package uleb128

import (
	"strings"
	"testing"
)

func TestWriteCoverageToBuffer(t *testing.T) {
	data, err := WriteCoverageToBuffer()
	if err != nil {
		if strings.Contains(err.Error(), "no meta-data available") {
			t.Skip("Binary not built with -cover; skipping test")
		}
		t.Fatalf("WriteCoverageToBuffer failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty buffer from WriteCoverageToBuffer")
	}
}
