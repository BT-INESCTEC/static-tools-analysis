package tool

import (
	"argus-benchmark/internal/target"
)

type Result struct {
	ExecutionTime float64
	Metrics       map[string]float64 // avg_cpu, peak_memory, etc.
	Timestamp     string
	Error         error
}

type Tool interface {
	Name() string
	Run(t target.Target, outputDir string) (*Result, error)
}
