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
		funcNameIdx, _ := reader.ReadULEB()
		fileNameIdx, _ := reader.ReadULEB()
		startLine, _ := reader.ReadULEB()
		endLine, _ := reader.ReadULEB()
		numCounters, _ := reader.ReadULEB()

		funcName := strings[int(funcNameIdx)]
		fileName := strings[int(fileNameIdx)]

		for i := 0; i < int(numCounters); i++ {
			entry := MetaEntry{
				FilePath:  fileName,
				FuncName:  funcName,
				LineStart: uint32(startLine), // assuming all counters span full func for now
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
