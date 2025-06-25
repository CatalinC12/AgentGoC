package lcovclient

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type LcovClient struct {
	AgentAddress  string
	FlushEndpoint string
	DumpDir       string
	DumpIndex     int
}

func NewLcovClient(agentAddr string, flushPath string, dumpDir string) (*LcovClient, error) {
	if err := os.MkdirAll(dumpDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create dump directory: %w", err)
	}
	return &LcovClient{
		AgentAddress:  agentAddr,
		FlushEndpoint: flushPath,
		DumpDir:       dumpDir,
		DumpIndex:     0,
	}, nil
}

func (lc *LcovClient) FetchCoverage(reset bool) error {
	if reset {
		// Reset dump dir by removing all files
		dirEntries, err := os.ReadDir(lc.DumpDir)
		if err != nil {
			return fmt.Errorf("failed to list dump directory: %w", err)
		}
		for _, entry := range dirEntries {
			_ = os.Remove(filepath.Join(lc.DumpDir, entry.Name()))
		}
		lc.DumpIndex = 0
	}

	// Trigger flush on the agent's instrumented app
	resp, err := http.Get(lc.FlushEndpoint)
	if err != nil {
		return fmt.Errorf("failed to flush coverage from app: %w", err)
	}
	_ = resp.Body.Close()

	// Connect to agent over TCP to fetch LCOV text
	conn, err := net.Dial("tcp", lc.AgentAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to agent: %w", err)
	}
	defer conn.Close()

	// Save LCOV content to file
	lcovFilePath := filepath.Join(lc.DumpDir, strconv.Itoa(lc.DumpIndex)+".lcov")
	file, err := os.Create(lcovFilePath)
	if err != nil {
		return fmt.Errorf("failed to create LCOV file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, conn); err != nil {
		return fmt.Errorf("failed to copy LCOV data to file: %w", err)
	}

	lc.DumpIndex++
	return nil
}
