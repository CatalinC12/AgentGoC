package src

import (
	"bytes"
	"fmt"
	"runtime/coverage"
)

// directly from the Go runtime’s in-memory tables.
func WriteCoverageToBuffers() (metaBuf, counterBuf []byte, err error) {
	var meta bytes.Buffer
	var counters bytes.Buffer

	if err = coverage.WriteMeta(&meta); err != nil {
		return nil, nil, fmt.Errorf("failed to write .covmeta: %w", err)
	}
	if err = coverage.WriteCounters(&counters); err != nil {
		return nil, nil, fmt.Errorf("failed to write .covcounters: %w", err)
	}
	return meta.Bytes(), counters.Bytes(), nil
}
