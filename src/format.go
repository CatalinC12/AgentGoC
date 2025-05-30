package uleb128

import (
	"fmt"
	"strings"
)

// EmitLcov converts decoded coverage data into a valid LCOV format string.
func EmitLcov(meta []MetaEntry, counters []uint64) (string, error) {
	var b strings.Builder
	files := make(map[string][]MetaEntry)

	// Group coverage entries by file
	for _, entry := range meta {
		files[entry.FilePath] = append(files[entry.FilePath], entry)
	}

	for file, entries := range files {
		b.WriteString(fmt.Sprintf("SF:%s\n", file)) // Source File

		funcLines := make(map[string]int)

		for _, entry := range entries {
			if entry.CounterID >= len(counters) {
				continue
			}

			count := counters[entry.CounterID]

			// FN:line,name
			if _, exists := funcLines[entry.FuncName]; !exists {
				b.WriteString(fmt.Sprintf("FN:%d,%s\n", entry.LineStart, entry.FuncName))
				funcLines[entry.FuncName] = int(entry.LineStart)
			}

			// FNDA:count,name
			b.WriteString(fmt.Sprintf("FNDA:%d,%s\n", count, entry.FuncName))

			// DA:line,count
			b.WriteString(fmt.Sprintf("DA:%d,%d\n", entry.LineStart, count))
		}

		// Function summary
		b.WriteString("FNF:")
		b.WriteString(fmt.Sprintf("%d\n", len(funcLines))) // number of functions

		b.WriteString("FNH:")
		covered := 0
		for _, entry := range entries {
			if entry.CounterID < len(counters) && counters[entry.CounterID] > 0 {
				covered++
			}
		}
		b.WriteString(fmt.Sprintf("%d\n", covered)) // number of hit functions

		b.WriteString("end_of_record\n")
	}

	return b.String(), nil
}
