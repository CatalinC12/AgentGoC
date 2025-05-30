package uleb128

import (
	"strings"
	"testing"
)

func TestEmitLcov(t *testing.T) {
	meta := []MetaEntry{
		{FilePath: "main.go", FuncName: "Foo", LineStart: 10, LineEnd: 20, CounterID: 0},
		{FilePath: "main.go", FuncName: "Bar", LineStart: 30, LineEnd: 40, CounterID: 1},
	}
	counters := []uint64{5, 0}

	output, err := EmitLcov(meta, counters)
	if err != nil {
		t.Fatalf("EmitLcov failed: %v", err)
	}

	// Basic structure check
	if !strings.Contains(output, "SF:main.go") ||
		!strings.Contains(output, "FN:10,Foo") ||
		!strings.Contains(output, "FNDA:5,Foo") ||
		!strings.Contains(output, "DA:10,5") ||
		!strings.Contains(output, "FN:30,Bar") ||
		!strings.Contains(output, "FNDA:0,Bar") ||
		!strings.Contains(output, "DA:30,0") {
		t.Errorf("LCOV output missing expected content:\n%s", output)
	}
}
