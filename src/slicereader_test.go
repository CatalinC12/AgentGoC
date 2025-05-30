package uleb128

import (
	"testing"
)

func TestSliceReader_ReadByte(t *testing.T) {
	data := []byte{0x01, 0x02}
	reader := NewSliceReader(data)

	b, err := reader.ReadByte()
	if err != nil || b != 0x01 {
		t.Errorf("Expected 0x01, got %x, err: %v", b, err)
	}

	b, err = reader.ReadByte()
	if err != nil || b != 0x02 {
		t.Errorf("Expected 0x02, got %x, err: %v", b, err)
	}

	_, err = reader.ReadByte()
	if err == nil {
		t.Error("Expected error on read past end")
	}
}

func TestSliceReader_ReadULEB(t *testing.T) {
	// 300 in ULEB128: 0xAC 0x02
	data := []byte{0xAC, 0x02}
	reader := NewSliceReader(data)

	val, err := reader.ReadULEB()
	if err != nil {
		t.Fatalf("Failed to decode ULEB128: %v", err)
	}
	if val != 300 {
		t.Errorf("Expected 300, got %d", val)
	}
}

func TestSliceReader_EOF(t *testing.T) {
	reader := NewSliceReader([]byte{0x01})
	_, _ = reader.ReadByte()
	if !reader.EOF() {
		t.Error("Expected EOF after reading last byte")
	}
}
