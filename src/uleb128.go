package uleb128

import (
	"errors"
	"io"
)

// ReadULEB128 reads a ULEB128-encoded unsigned integer from any io.ByteReader.
func ReadULEB128(r io.ByteReader) (uint64, error) {
	var result uint64
	var shift uint

	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= uint64(b&0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7

		// Prevent overflow
		if shift >= 64 {
			return 0, errors.New("ULEB128 overflow")
		}
	}
	return result, nil
}
