package src

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

// MetaEntry represents a single counter's mapping to source code.
type MetaEntry struct {
	FilePath  string
	FuncName  string
	LineStart uint32
	LineEnd   uint32
	CounterID int
}

func DecodeMeta(data []byte) ([]MetaEntry, error) {
	if len(data) < 80 {
		return nil, fmt.Errorf("data too short to be valid .covmeta")
	}

	// Read binary header
	var hdr metaFileHeader
	err := binary.Read(bytes.NewReader(data[:80]), binary.LittleEndian, &hdr)
	if err != nil {
		return nil, fmt.Errorf("failed to read covmeta header: %w", err)
	}
	if hdr.Magic != [4]byte{0x00, 'c', 'v', 'm'} {
		return nil, fmt.Errorf("invalid magic header: %q", hdr.Magic)
	}

	// Read package offset table
	offsetsStart := 80
	offsetsEnd := offsetsStart + int(hdr.NumPackages)*8
	if offsetsEnd > len(data) {
		return nil, fmt.Errorf("offset table out of bounds")
	}
	offsets := make([]uint64, hdr.NumPackages)
	err = binary.Read(bytes.NewReader(data[offsetsStart:offsetsEnd]), binary.LittleEndian, &offsets)
	if err != nil {
		return nil, fmt.Errorf("failed to read package offsets: %w", err)
	}

	// Read package length table
	lengthsStart := offsetsEnd
	lengthsEnd := lengthsStart + int(hdr.NumPackages)*8
	if lengthsEnd > len(data) {
		return nil, fmt.Errorf("length table out of bounds")
	}
	lengths := make([]uint64, hdr.NumPackages)
	err = binary.Read(bytes.NewReader(data[lengthsStart:lengthsEnd]), binary.LittleEndian, &lengths)
	if err != nil {
		return nil, fmt.Errorf("failed to read package lengths: %w", err)
	}

	// Read and decode shared string table
	strTabStart := int(hdr.StrTabOffset)
	strTabEnd := strTabStart + int(hdr.StrTabLength)
	if strTabEnd > len(data) {
		return nil, fmt.Errorf("string table out of bounds")
	}
	strReader := NewSliceReader(data[strTabStart:strTabEnd])
	strings, err := parseStringTable(strReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse string table: %w", err)
	}

	log.Printf("[Agent] String table contains %d entries", len(strings))

	//Decode each package blob
	var allEntries []MetaEntry
	for i := 0; i < int(hdr.NumPackages); i++ {
		off := int(offsets[i])
		end := off + int(lengths[i])
		if end > len(data) {
			return nil, fmt.Errorf("package %d metadata out of bounds", i)
		}
		pkgBlob := data[off:end]

		entries, err := decodePackageBlob(pkgBlob, strings)
		if err != nil {
			return nil, fmt.Errorf("error decoding package %d: %w", i, err)
		}
		allEntries = append(allEntries, entries...)
	}

	return allEntries, nil
}

func decodePackageBlob(data []byte, strings []string) ([]MetaEntry, error) {
	reader := NewSliceReader(data)

	// Skip package-level header (version, mode, granularity)
	_, _ = reader.ReadULEB()
	_, _ = reader.ReadULEB()
	_, _ = reader.ReadULEB()

	var entries []MetaEntry
	counterID := 0

	for !reader.EOF() {
		funcNameIdx, err := reader.ReadULEB()
		if err != nil || int(funcNameIdx) >= len(strings) {
			return nil, fmt.Errorf("invalid func name index")
		}
		fileIdx, err := reader.ReadULEB()
		if err != nil || int(fileIdx) >= len(strings) {
			return nil, fmt.Errorf("invalid file index")
		}
		startLine, _ := reader.ReadULEB()
		endLine, _ := reader.ReadULEB()
		numCounters, _ := reader.ReadULEB()

		for i := 0; i < int(numCounters); i++ {
			entries = append(entries, MetaEntry{
				FilePath:  strings[fileIdx],
				FuncName:  strings[funcNameIdx],
				LineStart: uint32(startLine),
				LineEnd:   uint32(endLine),
				CounterID: counterID,
			})
			counterID++
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

// metaFileHeader matches Go's internal MetaFileHeader layout.
type metaFileHeader struct {
	Magic        [4]byte
	Version      uint32
	TotalLen     uint64
	NumPackages  uint64
	Hash         [16]byte
	StrTabOffset uint32
	StrTabLength uint32
	CMode        uint8
	CGran        uint8
	_            [6]byte
}
