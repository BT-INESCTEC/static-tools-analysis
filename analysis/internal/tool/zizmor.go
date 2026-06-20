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
	cmd := exec.Command("zizmor",
		"--format", "sarif",
		t.Path(),
	)

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
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
		Stdout:        stdoutBuf.String(),
		Stderr:        stderrBuf.String(),
	}, nil
}

func (z *Zizmor) WriteSARIF(_ target.Target, outputDir string, r *Result) error {
	sarifPath := filepath.Join(outputDir, "zizmor.sarif")
	return os.WriteFile(sarifPath, []byte(r.Stdout), 0644)
}
