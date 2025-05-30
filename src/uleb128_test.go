package uleb128

import (
	"bytes"
	"testing"
)

func TestReadULEB128(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint64
	}{
		{"SingleByte127", []byte{0x7F}, 127},
		{"TwoBytes128", []byte{0x80, 0x01}, 128},
		{"TwoBytes300", []byte{0xAC, 0x02}, 300},
		{"Zero", []byte{0x00}, 0},
		{"LargeNumber", []byte{0xFF, 0xFF, 0x7F}, 2097151}, // 0x1FFFFF
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			val, err := ReadULEB128(reader)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, val)
			}
		})
	}
}
