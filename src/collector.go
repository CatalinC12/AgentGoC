package uleb128

import (
	"bytes"
	"fmt"
	"runtime/coverage"
)

// WriteCoverageToBuffer collects Go runtime coverage data and returns it as a byte slice.
func WriteCoverageToBuffer() ([]byte, error) {
	var output bytes.Buffer

	// Write coverage metadata (.covmeta)
	err := coverage.WriteMeta(&output)
	if err != nil {
		return nil, fmt.Errorf("failed to write .covmeta: %w", err)
	}

	// Write coverage counters (.covcounters)
	err = coverage.WriteCounters(&output)
	if err != nil {
		return nil, fmt.Errorf("failed to write .covcounters: %w", err)
	}

	return output.Bytes(), nil
}
