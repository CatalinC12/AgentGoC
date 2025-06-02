package src

import (
	"bytes"
	"testing"
)

// Minimal ULEB128 encoder (if not already imported)
func encodeULEB128(value int) []byte {
	var buf []byte
	for {
		b := byte(value & 0x7F)
		value >>= 7
		if value != 0 {
			b |= 0x80
		}
		buf = append(buf, b)
		if value == 0 {
			break
		}
	}
	return buf
}

// Helper to manually construct valid .covmeta and .covcounters
func TestAgentLCOV_InternalPipeline(t *testing.T) {
	var metaBuf bytes.Buffer

	// ---------- Simulate .covmeta ----------
	// String table: ["main", "main.go"]
	metaBuf.Write(encodeULEB128(2)) // string count
	metaBuf.Write(encodeULEB128(4)) // len("main")
	metaBuf.WriteString("main")     // "main"
	metaBuf.Write(encodeULEB128(7)) // len("main.go")
	metaBuf.WriteString("main.go")  // "main.go"

	// Function record
	metaBuf.Write(encodeULEB128(0))  // funcNameIdx -> "main"
	metaBuf.Write(encodeULEB128(1))  // fileNameIdx -> "main.go"
	metaBuf.Write(encodeULEB128(10)) // start line
	metaBuf.Write(encodeULEB128(20)) // end line
	metaBuf.Write(encodeULEB128(1))  // numCounters

	// ---------- Simulate .covcounters ----------
	// .covcounters format: ULEB(functionID) + ULEB(count)
	var counterBuf bytes.Buffer        // funcID = 0
	counterBuf.Write(encodeULEB128(5)) // counter = 5

	// ---------- Decode and Emit ----------
	metaEntries, err := DecodeMeta(metaBuf.Bytes())
	if err != nil {
		t.Fatalf("DecodeMeta failed: %v", err)
	}

	counters, err := DecodeCounters(counterBuf.Bytes())
	if err != nil {
		t.Fatalf("DecodeCounters failed: %v", err)
	}

	lcov, err := EmitLcov(metaEntries, counters)
	if err != nil {
		t.Fatalf("EmitLcov failed: %v", err)
	}

	if !bytes.Contains([]byte(lcov), []byte("SF:main.go")) {
		t.Errorf("Expected LCOV to include source file, got:\n%s", lcov)
	}

	t.Logf("Generated LCOV:\n%s", lcov)

	if !bytes.Contains([]byte(lcov), []byte("DA:10,5")) {
		t.Errorf("Expected LCOV to include DA entry for line 10, got:\n%s", lcov)
	}
}
