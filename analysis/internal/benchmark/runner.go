package benchmark

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"argus-benchmark/internal/config"
	"argus-benchmark/internal/monitor"
	"argus-benchmark/internal/results"
	"argus-benchmark/internal/target"
	"argus-benchmark/internal/tool"
)

type Runner struct {
	tools       []tool.Tool
	rawDstatDir string
	sarifDir    string
}

func NewRunner(tools []tool.Tool) *Runner {
	return &Runner{
		tools:       tools,
		rawDstatDir: filepath.Join(config.ResultsDir, "raw_dstat"),
		sarifDir:    filepath.Join(config.ResultsDir, "sarif_outputs"),
	}
}

func (r *Runner) Run(targets []target.Target) error {
	_ = os.MkdirAll(r.rawDstatDir, 0755)
	_ = os.MkdirAll(r.sarifDir, 0755)
	for _, t := range targets {
		_ = os.MkdirAll(filepath.Join(r.sarifDir, t.Name()), 0755)
	}

	if config.EnableDstat {
		log.Printf("\n📊 Collecting baseline statistics (%v idle)...", config.BaselineDuration)
		baseline, err := monitor.CollectBaseline(r.rawDstatDir)
		if err != nil {
			log.Printf("  ⚠️  Failed to collect baseline: %v", err)
		} else {
			baselineFile := filepath.Join(config.ResultsDir, "baseline_results.csv")
			if err := monitor.WriteBaselineCSV(baselineFile, baseline); err != nil {
				log.Printf("  ⚠️  Failed to write baseline CSV: %v", err)
			} else {
				log.Printf("  ✅ Baseline collected. Results saved to: %s", baselineFile)
			}
		}
	}

	writers := make(map[string]*results.Writer)
	for _, t := range r.tools {
		w, err := results.NewWriter(t.Name(), config.ResultsDir)
		if err != nil {
			return fmt.Errorf("failed to create writer for %s: %w", t.Name(), err)
		}
		writers[t.Name()] = w
		defer w.Close()
	}

	totalRuns := len(targets) * config.RunsPerWorkflow * len(r.tools)
	currentRun := 0

	for _, tgt := range targets {
		log.Printf("\n📁 Testing workflow: %s", tgt.Name())
		outputDir := filepath.Join(r.sarifDir, tgt.Name())

		for run := 1; run <= config.RunsPerWorkflow; run++ {
			for _, tl := range r.tools {
				currentRun++
				log.Printf("  [%d/%d] Run %d/%d - %s", currentRun, totalRuns, run, config.RunsPerWorkflow, tl.Name())

				dstat := monitor.NewDstat(r.rawDstatDir, fmt.Sprintf("%s_%s_run%d", tgt.Name(), tl.Name(), run))
				if err := dstat.Start(); err != nil {
					log.Printf("    ⚠️  Failed to start dstat: %v", err)
				}

				res, err := tl.Run(tgt, outputDir)
				if err != nil {
					log.Printf("    ❌ %s Error: %v", tl.Name(), err)
					_ = dstat.Stop()
					continue
				}

				if err := dstat.Stop(); err != nil {
					log.Printf("    ⚠️  Failed to stop dstat: %v", err)
				}

				var metrics *monitor.Metrics
				if config.EnableDstat {
					m, err := dstat.Parse()
					if err != nil {
						log.Printf("    ⚠️  Failed to parse dstat: %v", err)
					} else {
						metrics = m
					}
				}

				if w, ok := writers[tl.Name()]; ok {
					if err := w.Write(tgt, run, res, metrics); err != nil {
						log.Printf("    ⚠️  Failed to write result: %v", err)
					}
				}

				log.Printf("    ✅ %s completed in %.2fs", tl.Name(), res.ExecutionTime)
			}

			time.Sleep(2 * time.Second)
		}
	}

	log.Printf("\n✨ Benchmarking complete!")
	for _, tl := range r.tools {
		log.Printf("   %s results saved to: %s", tl.Name(), filepath.Join(config.ResultsDir, fmt.Sprintf("benchmark_results_%s.csv", tl.Name())))
	}
	log.Printf("   SARIF outputs saved to: %s\n", r.sarifDir)
	return nil
}
