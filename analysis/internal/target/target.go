package target

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"argus-benchmark/internal/config"
)

type Target interface {
	Name() string
	Path() string
}

type FileTarget struct {
	name string
	path string
}

func (f FileTarget) Name() string { return f.name }
func (f FileTarget) Path() string { return f.path }

func Discover(workflowsDir, ignoreFile string) ([]Target, error) {
	ignores, err := config.LoadIgnoreList(ignoreFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load ignore list: %w", err)
	}
	ignoreSet := make(map[string]struct{}, len(ignores))
	for _, ig := range ignores {
		ignoreSet[ig] = struct{}{}
	}

	entries, err := os.ReadDir(workflowsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflows dir: %w", err)
	}

	var targets []Target
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}
		if _, ok := ignoreSet[name]; ok {
			continue
		}
		targets = append(targets, FileTarget{
			name: strings.TrimSuffix(name, filepath.Ext(name)),
			path: filepath.Join(workflowsDir, name),
		})
	}
	return targets, nil
}
