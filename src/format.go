package src

import (
	"fmt"
	"sort"
	"strings"
)

func EmitLcov(meta []MetaEntry, counters []uint64) (string, error) {
	var b strings.Builder
	files := make(map[string][]MetaEntry)

	// Group coverage entries by file
	for _, entry := range meta {
		files[entry.FilePath] = append(files[entry.FilePath], entry)
	}

	// Process files in sorted order for consistency
	var sortedFiles []string
	for file := range files {
		sortedFiles = append(sortedFiles, file)
	}
	sort.Strings(sortedFiles)

	for _, file := range sortedFiles {
		entries := files[file]
		b.WriteString(fmt.Sprintf("SF:%s\n", file)) // Source File

		funcLines := make(map[string]int)
		funcHits := make(map[string]uint64)
		seenLines := make(map[int]bool)

		totalLines := 0
		hitLines := 0

		for _, entry := range entries {
			if entry.CounterID >= len(counters) {
				continue
			}
			count := counters[entry.CounterID]
			line := int(entry.LineStart)

			// FN: only once per function
			if _, exists := funcLines[entry.FuncName]; !exists {
				b.WriteString(fmt.Sprintf("FN:%d,%s\n", entry.LineStart, entry.FuncName))
				funcLines[entry.FuncName] = line
			}

			// Count max hits for FNDA
			funcHits[entry.FuncName] += count

			// DA: emit once per line
			if !seenLines[line] {
				b.WriteString(fmt.Sprintf("DA:%d,%d\n", line, count))
				seenLines[line] = true
				totalLines++
				if count > 0 {
					hitLines++
				}
			}
		}

		// Emit FNDA after all counts collected
		for name, line := range funcLines {
			b.WriteString(fmt.Sprintf("FN:%d,%s\n", line, name)) // Ensure FN line is included
			b.WriteString(fmt.Sprintf("FNDA:%d,%s\n", funcHits[name], name))
		}

		b.WriteString(fmt.Sprintf("FNF:%d\n", len(funcLines)))         // Functions Found
		b.WriteString(fmt.Sprintf("FNH:%d\n", countNonZero(funcHits))) // Functions Hit

		b.WriteString(fmt.Sprintf("LF:%d\n", totalLines)) // Lines Found
		b.WriteString(fmt.Sprintf("LH:%d\n", hitLines))   // Lines Hit

		b.WriteString("end_of_record\n")
	}

	return b.String(), nil
}

func countNonZero(m map[string]uint64) int {
	count := 0
	for _, v := range m {
		if v > 0 {
			count++
		}
	}
	return count
}
