package uleb128

import (
	"fmt"
)

// MetaEntry represents a single counter's mapping to source code.
type MetaEntry struct {
	FilePath  string
	FuncName  string
	LineStart uint32
	LineEnd   uint32
	CounterID int
}

// DecodeMeta decodes Go's binary .covmeta coverage metadata format.
func DecodeMeta(data []byte) ([]MetaEntry, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("DecodeMeta: data too short (%d bytes)", len(data))
	}

	reader := NewSliceReader(data)

	// Parse string table (shared names)
	strings, err := parseStringTable(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse string table: %w", err)
	}

	var entries []MetaEntry
	counterID := 0

	// Parse function records
	for !reader.EOF() {
		funcNameIdx, err := reader.ReadULEB()
		if err != nil || int(funcNameIdx) >= len(strings) {
			return nil, fmt.Errorf("invalid funcName index: %v", funcNameIdx)
		}

		fileNameIdx, err := reader.ReadULEB()
		if err != nil || int(fileNameIdx) >= len(strings) {
			return nil, fmt.Errorf("invalid fileName index: %v", fileNameIdx)
		}

		startLine, err := reader.ReadULEB()
		if err != nil {
			return nil, fmt.Errorf("invalid startLine: %w", err)
		}

		endLine, err := reader.ReadULEB()
		if err != nil {
			return nil, fmt.Errorf("invalid endLine: %w", err)
		}

		numCounters, err := reader.ReadULEB()
		if err != nil {
			return nil, fmt.Errorf("invalid numCounters: %w", err)
		}

		funcName := strings[int(funcNameIdx)]
		fileName := strings[int(fileNameIdx)]

		for i := 0; i < int(numCounters); i++ {
			entry := MetaEntry{
				FilePath:  fileName,
				FuncName:  funcName,
				LineStart: uint32(startLine),
				LineEnd:   uint32(endLine),
				CounterID: counterID,
			}
			counterID++
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

func parseStringTable(r *SliceReader) ([]string, error) {
	var strings []string

	count, err := r.ReadULEB()
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(count); i++ {
		length, err := r.ReadULEB()
		if err != nil {
			return nil, err
		}

		strBytes := make([]byte, length)
		for j := uint64(0); j < length; j++ {
			b, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			strBytes[j] = b
		}

		strings = append(strings, string(strBytes))
	}

	return strings, nil
}
