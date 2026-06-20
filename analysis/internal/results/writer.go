package results

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"argus-benchmark/internal/config"
	"argus-benchmark/internal/monitor"
	"argus-benchmark/internal/target"
	"argus-benchmark/internal/tool"
)

type Writer struct {
	toolName string
	file     *os.File
	csv      *csv.Writer
}

func NewWriter(toolName, resultsDir string) (*Writer, error) {
	filePath := filepath.Join(resultsDir, fmt.Sprintf("benchmark_results_%s.csv", toolName))
	f, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create results file for %s: %w", toolName, err)
	}

	w := &Writer{
		toolName: toolName,
		file:     f,
		csv:      csv.NewWriter(f),
	}

	header := w.buildHeader()
	if err := w.csv.Write(header); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to write header for %s: %w", toolName, err)
	}
	w.csv.Flush()
	return w, nil
}

func (w *Writer) Write(t target.Target, runNumber int, res *tool.Result, metrics *monitor.Metrics) error {
	record := w.buildRecord(t, runNumber, res, metrics)
	if err := w.csv.Write(record); err != nil {
		return fmt.Errorf("failed to write record for %s: %w", w.toolName, err)
	}
	w.csv.Flush()
	return nil
}

func (w *Writer) Close() error {
	w.csv.Flush()
	return w.file.Close()
}

func (w *Writer) buildHeader() []string {
	header := []string{"workflow_file", "run_number", "execution_time_seconds"}
	if config.EnableDstat {
		header = append(header,
			"avg_cpu_percent", "peak_memory_mb",
			"avg_disk_read_kb", "avg_disk_write_kb",
			"avg_net_recv_kb", "avg_net_send_kb",
		)
	}
	header = append(header, "timestamp")
	return header
}

func (w *Writer) buildRecord(t target.Target, runNumber int, res *tool.Result, metrics *monitor.Metrics) []string {
	record := []string{
		t.Name(),
		strconv.Itoa(runNumber),
		fmt.Sprintf("%.3f", res.ExecutionTime),
	}
	if config.EnableDstat && metrics != nil {
		record = append(record,
			fmt.Sprintf("%.2f", metrics.AvgCPU),
			fmt.Sprintf("%.2f", metrics.PeakMemory),
			fmt.Sprintf("%.2f", metrics.AvgDiskRead),
			fmt.Sprintf("%.2f", metrics.AvgDiskWrite),
			fmt.Sprintf("%.2f", metrics.AvgNetRecv),
			fmt.Sprintf("%.2f", metrics.AvgNetSend),
		)
	}
	record = append(record, res.Timestamp)
	return record
}
