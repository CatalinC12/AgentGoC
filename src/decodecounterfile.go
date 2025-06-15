package src

import (
	"bytes"
	"fmt"
	"io"
)

// These values are typically ULEB128 encoded and reflect execution counts.
func DecodeCounters(data []byte) ([]uint64, error) {
	reader := bytes.NewReader(data)
	var counters []uint64

	for {
		value, err := readULEB128(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading ULEB128 counter: %w", err)
		}
		counters = append(counters, value)
	}

	return counters, nil
}

// readULEB128 reads a single unsigned LEB128-encoded integer from the reader.
func readULEB128(r io.ByteReader) (uint64, error) {
	var result uint64
	var shift uint

	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return result, nil
}
