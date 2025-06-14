package src

import (
	"strings"
	"testing"
)

func TestWriteCoverageToBuffers(t *testing.T) {
	metaBuf, counterBuf, err := WriteCoverageToBuffers()
	if err != nil {
		if strings.Contains(err.Error(), "no meta-data available") {
			t.Skip("Binary not built with -cover; skipping test")
		}
		t.Fatalf("WriteCoverageToBuffers failed: %v", err)
	}

	if len(metaBuf) == 0 {
		t.Error("Expected non-empty meta buffer from WriteCoverageToBuffers")
	}
	if len(counterBuf) == 0 {
		t.Error("Expected non-empty counters buffer from WriteCoverageToBuffers")
	}
}
