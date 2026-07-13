package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

type BundleVariant struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Script      string `json:"script"`
}

type BundleEntry struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Variants    []BundleVariant `json:"variants"`
}

type SampleData struct {
	Metadata string
	Payload  string
	ReadErr  error
}

func TestTemplateSamples(t *testing.T) {
	baseDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting working directory: %v", err)
	}
	repoDir := filepath.Dir(baseDir)

	scriptBytes, err := os.ReadFile(filepath.Join(baseDir, "eat.rice"))
	if err != nil {
		t.Fatalf("Error reading eat.rice: %v", err)
	}

	bundleBytes, err := os.ReadFile(filepath.Join(repoDir, "templates", "bundle.json"))
	if err != nil {
		t.Fatalf("Error reading templates/bundle.json: %v", err)
	}

	var bundle []BundleEntry
	if err := json.Unmarshal(bundleBytes, &bundle); err != nil {
		t.Fatalf("Error parsing templates/bundle.json: %v", err)
	}

	indexBytes, err := os.ReadFile(filepath.Join(repoDir, "test-samples", "index.json"))
	if err != nil {
		t.Fatalf("Error reading test-samples/index.json: %v", err)
	}

	var entries []IndexEntry
	if err := json.Unmarshal(indexBytes, &entries); err != nil {
		t.Fatalf("Error parsing test-samples/index.json: %v", err)
	}

	sampleCache := make(map[string]SampleData, len(entries))
	for _, entry := range entries {
		key := entry.Protocol + "/" + entry.Sample
		if _, exists := sampleCache[key]; exists {
			continue
		}

		var sd SampleData
		metaBytes, err := os.ReadFile(filepath.Join(repoDir, "test-samples", filepath.FromSlash(entry.Metadata)))
		if err != nil {
			sd.ReadErr = fmt.Errorf("failed to read metadata: %w", err)
			sampleCache[key] = sd
			continue
		}
		sd.Metadata = string(metaBytes)

		if entry.Payload != nil {
			payloadBytes, err := os.ReadFile(filepath.Join(repoDir, "test-samples", filepath.FromSlash(*entry.Payload)))
			if err != nil {
				sd.ReadErr = fmt.Errorf("failed to read payload: %w", err)
				sampleCache[key] = sd
				continue
			}
			sd.Payload = string(payloadBytes)
		}

		sampleCache[key] = sd
	}

	for _, tmpl := range bundle {
		for _, variant := range tmpl.Variants {
			t.Run(tmpl.Name+"/"+variant.Name, func(t *testing.T) {
				t.Parallel()

				scriptTxt := string(scriptBytes)
				scriptTxt = strings.ReplaceAll(scriptTxt, "#{{__script__}}", variant.Script)

				tokens, err := frontend.Tokenize(scriptTxt)
				if err != nil {
					t.Fatalf("[%s/%s] Error tokenizing: %v", tmpl.Name, variant.Name, err)
				}

				parser := frontend.NewParser(tokens)
				ast := parser.Parse()
				if len(parser.Errors()) > 0 {
					for _, e := range parser.Errors() {
						t.Errorf("[%s/%s] Parse error: %s", tmpl.Name, variant.Name, e)
					}
					t.FailNow()
				}

				for _, entry := range entries {
					t.Run(entry.Protocol+"/"+entry.Sample, func(t *testing.T) {
						t.Parallel()

						sd := sampleCache[entry.Protocol+"/"+entry.Sample]
						if sd.ReadErr != nil {
							t.Fatalf("sample read error: %v", sd.ReadErr)
						}

						runCfg := conf.NewDefaultRunConfig().
							DefineConstant(values.Identifier("_metadata"), values.String(sd.Metadata)).
							DefineConstant(values.Identifier("_payload"), values.String(sd.Payload))

						it := exec.NewInterpreter(conf.NewDefaultEnvConfig())
						_, runErr := it.Interpret(context.Background(), ast, runCfg)

						if runErr != nil {
							var re exec.RuntimeError
							if errors.As(runErr, &re) {
								t.Errorf("%s", re.Stacktrace())
							} else {
								t.Errorf("%v", runErr)
							}
						}
					})
				}
			})
		}
	}
}
