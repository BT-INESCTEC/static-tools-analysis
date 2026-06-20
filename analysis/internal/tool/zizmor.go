package tool

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"argus-benchmark/internal/target"
)

type Zizmor struct{}

func (z *Zizmor) Name() string { return "zizmor" }

func (z *Zizmor) Run(t target.Target, outputDir string) (*Result, error) {
	outputFile := filepath.Join(outputDir, "zizmor.sarif")
	stdoutFile, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer stdoutFile.Close()

	cmd := exec.Command("zizmor",
		"--format", "sarif",
		t.Path(),
	)
	cmd.Stdout = stdoutFile

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	start := time.Now()
	if err := cmd.Run(); err != nil {
		stderrOutput := stderrBuf.String()
		if stderrOutput != "" {
			return nil, fmt.Errorf("zizmor failed: %w (stderr: %s)", err, stderrOutput[:min(len(stderrOutput), 200)])
		}
		return nil, fmt.Errorf("zizmor failed: %w", err)
	}
	elapsed := time.Since(start).Seconds()

	return &Result{
		ExecutionTime: elapsed,
		Timestamp:     time.Now().Format("2006-01-02_15-04-05"),
	}, nil
}
