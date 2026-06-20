package main

import (
	"log"

	"argus-benchmark/internal/benchmark"
	"argus-benchmark/internal/config"
	"argus-benchmark/internal/target"
	"argus-benchmark/internal/tool"
)

func main() {
	log.Println("Starting Argus Benchmarking Suite")

	targets, err := target.Discover(config.WorkflowsDir, config.IgnoreFilePath)
	if err != nil {
		log.Fatalf("Failed to discover targets: %v", err)
	}

	log.Printf("Discovered %d workflows", len(targets))
	if config.EnableDstat {
		log.Println("dstat resource monitoring: ENABLED")
	} else {
		log.Println("dstat resource monitoring: DISABLED (timing only)")
	}
	log.Printf("Testing with %d runs each using %d tools", config.RunsPerWorkflow, len(tool.Registry))

	var tools []tool.Tool
	for _, t := range tool.Registry {
		tools = append(tools, t)
	}

	runner := benchmark.NewRunner(tools)
	if err := runner.Run(targets); err != nil {
		log.Fatalf("Benchmark failed: %v", err)
	}
}
