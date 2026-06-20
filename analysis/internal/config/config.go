package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	RunsPerWorkflow = 3
	EnableDstat     = true
	DstatPreDelay   = 1 * time.Second
	DstatPostDelay  = 1 * time.Second
	BaselineDuration = 10 * time.Second
)

var (
	WorkflowsDir    = filepath.Join("..", "test_workflows")
	IgnoreFilePath  = filepath.Join(WorkflowsDir, ".benchmarkignore")
	ResultsDir      = "results"
)

func LoadIgnoreList(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var ignores []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ignores = append(ignores, line)
	}
	return ignores, scanner.Err()
}
