package uleb128

import (
	"testing"
)

func TestDecodeCounters(t *testing.T) {
	input := []byte{0x7F, 0x80, 0x01, 0xAC, 0x02} // ULEB128: 127, 128, 300
	expected := []uint64{127, 128, 300}

	counters, err := DecodeCounters(input)
	if err != nil {
		t.Fatalf("DecodeCounters returned error: %v", err)
	}

	if len(counters) != len(expected) {
		t.Fatalf("Expected %d counters, got %d", len(expected), len(counters))
	}

	for i := range counters {
		if counters[i] != expected[i] {
			t.Errorf("Counter %d: expected %d, got %d", i, expected[i], counters[i])
		}
	}
}
