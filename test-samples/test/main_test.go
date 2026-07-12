package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/anhcraft/rice/exec"
	"github.com/anhcraft/rice/exec/conf"
	"github.com/anhcraft/rice/exec/types/values"
	"github.com/anhcraft/rice/frontend"
)

type IndexEntry struct {
	Sample   string  `json:"sample"`
	Protocol string  `json:"protocol"`
	Metadata string  `json:"metadata"`
	Payload  *string `json:"payload"`
}

type TestResult struct {
	Sample   string `json:"sample"`
	Protocol string `json:"protocol"`
	Passed   bool   `json:"passed"`
	Error    string `json:"error,omitempty"`
}

func TestSamples(t *testing.T) {
	baseDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting working directory: %v", err)
	}
	testSamplesDir := filepath.Dir(baseDir)

	scriptBytes, err := os.ReadFile(filepath.Join(baseDir, "test.rice"))
	if err != nil {
		t.Fatalf("Error reading test.rice: %v", err)
	}

	tokens, err := frontend.Tokenize(string(scriptBytes))
	if err != nil {
		t.Fatalf("Error tokenizing: %v", err)
	}

	parser := frontend.NewParser(tokens)
	ast := parser.Parse()
	if len(parser.Errors()) > 0 {
		for _, e := range parser.Errors() {
			t.Errorf("Parse error: %s", e)
		}
		t.FailNow()
	}

	indexBytes, err := os.ReadFile(filepath.Join(testSamplesDir, "index.json"))
	if err != nil {
		t.Fatalf("Error reading index.json: %v", err)
	}

	var entries []IndexEntry
	if err := json.Unmarshal(indexBytes, &entries); err != nil {
		t.Fatalf("Error parsing index.json: %v", err)
	}

	var (
		resultsMu sync.Mutex
		results   []TestResult
		wg        sync.WaitGroup
	)

	for _, entry := range entries {
		wg.Add(1)
		go func(e IndexEntry) {
			defer wg.Done()

			metaBytes, err := os.ReadFile(filepath.Join(testSamplesDir, filepath.FromSlash(e.Metadata)))
			if err != nil {
				resultsMu.Lock()
				results = append(results, TestResult{
					Sample: e.Sample, Protocol: e.Protocol, Passed: false,
					Error: fmt.Sprintf("failed to read metadata: %v", err),
				})
				resultsMu.Unlock()
				return
			}

			var payloadStr string
			if e.Payload != nil {
				payloadBytes, err := os.ReadFile(filepath.Join(testSamplesDir, filepath.FromSlash(*e.Payload)))
				if err != nil {
					resultsMu.Lock()
					results = append(results, TestResult{
						Sample: e.Sample, Protocol: e.Protocol, Passed: false,
						Error: fmt.Sprintf("failed to read payload: %v", err),
					})
					resultsMu.Unlock()
					return
				}
				payloadStr = string(payloadBytes)
			}

			runCfg := conf.NewDefaultRunConfig().
				DefineConstant(values.Identifier("_metadata"), values.String(string(metaBytes))).
				DefineConstant(values.Identifier("_payload"), values.String(payloadStr))

			it := exec.NewInterpreter(conf.NewDefaultEnvConfig())
			_, runErr := it.Interpret(context.Background(), ast, runCfg)

			resultsMu.Lock()
			defer resultsMu.Unlock()

			if runErr != nil {
				var re exec.RuntimeError
				if errors.As(runErr, &re) {
					results = append(results, TestResult{
						Sample: e.Sample, Protocol: e.Protocol, Passed: false,
						Error: re.Stacktrace(),
					})
				} else {
					results = append(results, TestResult{
						Sample: e.Sample, Protocol: e.Protocol, Passed: false,
						Error: runErr.Error(),
					})
				}
			} else {
				results = append(results, TestResult{
					Sample: e.Sample, Protocol: e.Protocol, Passed: true,
				})
			}
		}(entry)
	}

	wg.Wait()

	for _, r := range results {
		t.Run(r.Protocol+"/"+r.Sample, func(t *testing.T) {
			if !r.Passed {
				t.Errorf("%s", r.Error)
			}
		})
	}

	passed, failed := 0, 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}
	t.Logf("%d passed, %d failed, %d total", passed, failed, len(results))

	if failed > 0 {
		t.Fail()
	}
}
