package src

import (
	"bytes"
	"fmt"
	"runtime/coverage"
)

// WriteCoverageToBuffers collects Go runtime coverage data and returns separate buffers.
func WriteCoverageToBuffers() (metaBuf, counterBuf []byte, err error) {
	var meta bytes.Buffer
	var counters bytes.Buffer

	err = coverage.WriteMeta(&meta)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write .covmeta: %w", err)
	}

	err = coverage.WriteCounters(&counters)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write .covcounters: %w", err)
	}

	return meta.Bytes(), counters.Bytes(), nil
}
