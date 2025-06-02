package src

import (
	"bytes"
	"testing"
)

// Helper to encode ULEB128 values
func EncodeULEB128(value uint64) []byte {
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

// Helper to build test metadata buffer for DecodeMeta
func encodeTestMetaData() []byte {
	var buf bytes.Buffer

	// String table: ["foo", "main.go"]
	buf.Write(EncodeULEB128(2)) // string count
	buf.Write(EncodeULEB128(3)) // "foo"
	buf.WriteString("foo")
	buf.Write(EncodeULEB128(7)) // "main.go"
	buf.WriteString("main.go")

	// Function record
	buf.Write(EncodeULEB128(0))  // funcNameIdx -> "foo"
	buf.Write(EncodeULEB128(1))  // fileNameIdx -> "main.go"
	buf.Write(EncodeULEB128(10)) // start line
	buf.Write(EncodeULEB128(20)) // end line
	buf.Write(EncodeULEB128(1))  // numCounters

	return buf.Bytes()
}

func TestDecodeMeta_ValidSingleFunction(t *testing.T) {
	data := encodeTestMetaData()

	meta, err := DecodeMeta(data)
	if err != nil {
		t.Fatalf("DecodeMeta failed: %v", err)
	}
	if len(meta) != 1 {
		t.Fatalf("Expected 1 meta entry, got %d", len(meta))
	}

	entry := meta[0]
	if entry.FuncName != "foo" || entry.FilePath != "main.go" || entry.LineStart != 10 || entry.LineEnd != 20 {
		t.Errorf("Incorrect MetaEntry decoded: %+v", entry)
	}
}
