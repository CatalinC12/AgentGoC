package src

import (
	"testing"
)

// helper to encode a string table and one function record
func encodeTestMetaData() []byte {
	data := []byte{}
	// String table: 2 strings ("foo", "main.go")
	data = append(data, 0x02) // count = 2
	data = append(data, 0x03) // len("foo")
	data = append(data, 'f', 'o', 'o')
	data = append(data, 0x07) // len("main.go")
	data = append(data, 'm', 'a', 'i', 'n', '.', 'g', 'o')

	// Function record: func="foo", file="main.go", lineStart=10, lineEnd=20, counters=1
	data = append(data, 0x00) // index of "foo"
	data = append(data, 0x01) // index of "main.go"
	data = append(data, 0x0A) // start line = 10
	data = append(data, 0x14) // end line = 20
	data = append(data, 0x01) // 1 counter

	return data
}

func TestDecodeMeta(t *testing.T) {
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
