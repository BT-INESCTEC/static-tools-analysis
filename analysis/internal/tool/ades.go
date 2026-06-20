package tool

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"argus-benchmark/internal/target"
)

// adesVersion must match the version pinned in go.mod
type Ades struct{}

func (a *Ades) Name() string { return "ades" }

func (a *Ades) Run(t target.Target, outputDir string) (*Result, error) {
	adesBin, err := findAdesBinary()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(adesBin, t.Path())

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start).Seconds()

	// ades exits 2 when violations are found — that's success for us
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok && exitErr.ExitCode() == 2 {
			// violations found — expected
		} else {
			stderrOutput := stderrBuf.String()
			if stderrOutput != "" {
				return nil, fmt.Errorf("ades failed: %w (stderr: %s)", runErr, stderrOutput[:min(len(stderrOutput), 200)])
			}
			return nil, fmt.Errorf("ades failed: %w", runErr)
		}
	}

	stdout := stdoutBuf.String()
	findings, err := parseAdesOutput(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ades output: %w", err)
	}

	sarifPath := filepath.Join(outputDir, "ades.sarif")
	if err := writeAdesSARIF(sarifPath, t, findings); err != nil {
		return nil, fmt.Errorf("failed to write ades SARIF: %w", err)
	}

	return &Result{
		ExecutionTime: elapsed,
		Timestamp:     time.Now().Format("2006-01-02_15-04-05"),
	}, nil
}

// findAdesBinary locates the ades executable.
func findAdesBinary() (string, error) {
	// 1. PATH lookup
	if p, err := exec.LookPath("ades"); err == nil {
		return p, nil
	}

	// 2. Common GOPATH locations
	for _, base := range []string{os.Getenv("GOPATH"), filepath.Join(os.Getenv("HOME"), "go")} {
		if base == "" {
			continue
		}
		p := filepath.Join(base, "bin", "ades")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("ades binary not found in PATH or GOPATH; run: go install github.com/ericcornelissen/ades/cmd/ades@v26.03.0")
}

type adesFinding struct {
	Job        string
	Step       string
	Expression string
	RuleID     string
}

var (
	totalRe  = regexp.MustCompile(`^Detected (\d+) violation\(s\) in "([^"]+)":`)
	jobRe    = regexp.MustCompile(`^\s*(\d+) in job "([^"]+)"\s*:\s*$`)
	stepRe   = regexp.MustCompile(`^\s+step "([^"]+)" contains "([^"]+)" \(([A-Z0-9]+)\)$`)
)

func parseAdesOutput(output string) ([]adesFinding, error) {
	output = strings.TrimSpace(output)
	if output == "Ok" {
		return nil, nil
	}

	var findings []adesFinding
	var currentJob string

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")

		if totalRe.MatchString(line) {
			continue
		}

		if m := jobRe.FindStringSubmatch(line); m != nil {
			currentJob = m[2]
			continue
		}

		if m := stepRe.FindStringSubmatch(line); m != nil {
			findings = append(findings, adesFinding{
				Job:        currentJob,
				Step:       m[1],
				Expression: m[2],
				RuleID:     m[3],
			})
			continue
		}
	}

	return findings, nil
}

func writeAdesSARIF(path string, t target.Target, findings []adesFinding) error {
	type artifactLocation struct {
		URI string `json:"uri"`
	}
	type physicalLocation struct {
		ArtifactLocation artifactLocation `json:"artifactLocation"`
	}
	type location struct {
		PhysicalLocation physicalLocation `json:"physicalLocation"`
	}
	type message struct {
		Text string `json:"text"`
	}
	type result struct {
		RuleID    string     `json:"ruleId"`
		Level     string     `json:"level"`
		Message   message    `json:"message"`
		Locations []location `json:"locations"`
	}
	type driver struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	type tool struct {
		Driver driver `json:"driver"`
	}
	type run struct {
		Tool    tool     `json:"tool"`
		Results []result `json:"results"`
	}
	type sarifOutput struct {
		Version string `json:"version"`
		Schema  string `json:"$schema"`
		Runs    []run  `json:"runs"`
	}

	var results []result
	for _, f := range findings {
		results = append(results, result{
			RuleID: f.RuleID,
			Level:  "warning",
			Message: message{
				Text: fmt.Sprintf(`step "%s" contains "%s"`, f.Step, f.Expression),
			},
			Locations: []location{{
				PhysicalLocation: physicalLocation{
					ArtifactLocation: artifactLocation{
						URI: t.Path(),
					},
				},
			}},
		})
	}

	out := sarifOutput{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []run{{
			Tool: tool{
				Driver: driver{
					Name:    "ades",
					Version: "v26.03.0",
				},
			},
			Results: results,
		}},
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
