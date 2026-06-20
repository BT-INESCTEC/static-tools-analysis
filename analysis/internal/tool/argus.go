package tool

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"argus-benchmark/internal/config"
	"argus-benchmark/internal/target"
)

type Argus struct{}

func (a *Argus) Name() string { return "argus" }

func (a *Argus) Run(t target.Target, outputDir string) (*Result, error) {
	args := []string{"run", "python3", "argus.py",
		"--mode", "file",
		"--file", t.Path(),
	}

	if config.SaveSARIF {
		outputFile, err := filepath.Abs(filepath.Join(outputDir, "argus.sarif"))
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for output file: %w", err)
		}
		args = append(args, "--output", outputFile)
	}

	cmd := exec.Command("poetry", args...)
	cmd.Dir = filepath.Join("..", "argus")

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
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
		Stdout:        stdoutBuf.String(),
		Stderr:        stderrBuf.String(),
	}, nil
}

// WriteSARIF is a no-op for argus: argus writes the SARIF file itself in Run
// when config.SaveSARIF is true.
func (a *Argus) WriteSARIF(_ target.Target, _ string, _ *Result) error {
	return nil
}
