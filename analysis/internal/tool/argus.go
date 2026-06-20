package tool

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"argus-benchmark/internal/target"
)

type Argus struct{}

func (a *Argus) Name() string { return "argus" }

func (a *Argus) Run(t target.Target, outputDir string) (*Result, error) {
	outputFile, err := filepath.Abs(filepath.Join(outputDir, "argus.sarif"))
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for output file: %w", err)
	}

	cmd := exec.Command("poetry", "run", "python3", "argus.py",
		"--mode", "file",
		"--file", t.Path(),
		"--output", outputFile,
	)
	cmd.Dir = filepath.Join("..", "argus")

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	start := time.Now()
	if err := cmd.Run(); err != nil {
		stderrOutput := stderrBuf.String()
		if stderrOutput != "" {
			return nil, fmt.Errorf("argus failed: %w (stderr: %s)", err, stderrOutput[:min(len(stderrOutput), 200)])
		}
		return nil, fmt.Errorf("argus failed: %w", err)
	}
	elapsed := time.Since(start).Seconds()

	return &Result{
		ExecutionTime: elapsed,
		Timestamp:     time.Now().Format("2006-01-02_15-04-05"),
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
