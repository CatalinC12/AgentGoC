package uleb128

import (
	"errors"
)

// SliceReader wraps a byte slice and provides ULEB128 decoding utilities.
type SliceReader struct {
	data []byte
	pos  int
}

// NewSliceReader creates a new reader from a byte slice.
func NewSliceReader(data []byte) *SliceReader {
	return &SliceReader{
		data: data,
		pos:  0,
	}
}

// ReadByte returns the next byte in the buffer.
func (r *SliceReader) ReadByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, errors.New("end of slice")
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

// ReadULEB reads a ULEB128-encoded unsigned integer.
func (r *SliceReader) ReadULEB() (uint64, error) {
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

// EOF returns true if all data has been read.
func (r *SliceReader) EOF() bool {
	return r.pos >= len(r.data)
}
