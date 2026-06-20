package monitor

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"argus-benchmark/internal/config"
)

type Metrics struct {
	AvgCPU       float64
	PeakMemory   float64
	AvgDiskRead  float64
	AvgDiskWrite float64
	AvgNetRecv   float64
	AvgNetSend   float64
}

type BaselineResult struct {
	Metrics   Metrics
	Timestamp string
}

type Dstat struct {
	cmd      *exec.Cmd
	pid      int
	outFile  string
}

func NewDstat(rawDstatDir, prefix string) *Dstat {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	outFile := filepath.Join(rawDstatDir, fmt.Sprintf("%s_%s.csv", prefix, timestamp))
	return &Dstat{outFile: outFile}
}

func (d *Dstat) Start() error {
	if !config.EnableDstat {
		return nil
	}
	d.cmd = exec.Command("dstat",
		"--time", "--cpu", "--mem", "--net", "--disk", "--swap",
		"--output", d.outFile)
	d.cmd.Stdout = nil
	d.cmd.Stderr = nil
	if err := d.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start dstat: %w", err)
	}
	d.pid = d.cmd.Process.Pid
	time.Sleep(config.DstatPreDelay)
	return nil
}

func (d *Dstat) Stop() error {
	if !config.EnableDstat || d.cmd == nil {
		return nil
	}
	if p, err := os.FindProcess(d.pid); err == nil {
		p.Kill()
	}
	d.cmd.Wait()
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (d *Dstat) Parse() (*Metrics, error) {
	if !config.EnableDstat {
		return nil, nil
	}
	return parseDstatOutput(d.outFile)
}

func CollectBaseline(rawDstatDir string) (*BaselineResult, error) {
	if !config.EnableDstat {
		return nil, nil
	}
	d := NewDstat(rawDstatDir, "baseline")
	if err := d.Start(); err != nil {
		return nil, err
	}
	time.Sleep(config.BaselineDuration)
	if err := d.Stop(); err != nil {
		return nil, err
	}
	metrics, err := d.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse baseline dstat: %w", err)
	}
	return &BaselineResult{
		Metrics:   *metrics,
		Timestamp: time.Now().Format("2006-01-02_15-04-05"),
	}, nil
}

func WriteBaselineCSV(filePath string, baseline *BaselineResult) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	header := []string{
		"avg_cpu_percent", "peak_memory_mb",
		"avg_disk_read_kb", "avg_disk_write_kb",
		"avg_net_recv_kb", "avg_net_send_kb",
		"timestamp",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	record := []string{
		fmt.Sprintf("%.2f", baseline.Metrics.AvgCPU),
		fmt.Sprintf("%.2f", baseline.Metrics.PeakMemory),
		fmt.Sprintf("%.2f", baseline.Metrics.AvgDiskRead),
		fmt.Sprintf("%.2f", baseline.Metrics.AvgDiskWrite),
		fmt.Sprintf("%.2f", baseline.Metrics.AvgNetRecv),
		fmt.Sprintf("%.2f", baseline.Metrics.AvgNetSend),
		baseline.Timestamp,
	}
	return w.Write(record)
}

func parseDstatOutput(filePath string) (*Metrics, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	dataStart := 6
	if len(records) <= dataStart {
		return nil, fmt.Errorf("no data found in dstat output")
	}

	var cpuValues, memValues, diskReadValues, diskWriteValues, netRecvValues, netSendValues []float64

	for i := dataStart; i < len(records); i++ {
		record := records[i]
		if len(record) < 14 {
			continue
		}

		if cpuUsr, err := parseFloat(record[1]); err == nil {
			if cpuSys, err := parseFloat(record[2]); err == nil {
				cpuValues = append(cpuValues, cpuUsr+cpuSys)
			}
		}

		if mem, err := parseFloat(record[6]); err == nil {
			memValues = append(memValues, mem/1024/1024)
		}

		if diskRead, err := parseFloat(record[12]); err == nil {
			diskReadValues = append(diskReadValues, diskRead/1024)
		}
		if diskWrite, err := parseFloat(record[13]); err == nil {
			diskWriteValues = append(diskWriteValues, diskWrite/1024)
		}

		if netRecv, err := parseFloat(record[10]); err == nil {
			netRecvValues = append(netRecvValues, netRecv/1024)
		}
		if netSend, err := parseFloat(record[11]); err == nil {
			netSendValues = append(netSendValues, netSend/1024)
		}
	}

	return &Metrics{
		AvgCPU:       average(cpuValues),
		PeakMemory:   max(memValues),
		AvgDiskRead:  average(diskReadValues),
		AvgDiskWrite: average(diskWriteValues),
		AvgNetRecv:   average(netRecvValues),
		AvgNetSend:   average(netSendValues),
	}, nil
}

func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	return strconv.ParseFloat(s, 64)
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}
