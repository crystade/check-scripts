package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type SampleEntry struct {
	Sample   string  `json:"sample"`
	Protocol string  `json:"protocol"`
	Metadata string  `json:"metadata"`
	Payload  *string `json:"payload"`
}

type BundledSample struct {
	Sample   string  `json:"sample"`
	Protocol string  `json:"protocol"`
	Metadata string  `json:"metadata"`
	Payload  *string `json:"payload"`
}

func main() {
	testSamplesDir := "test-samples"
	indexPath := filepath.Join(testSamplesDir, "index.json")

	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading index.json: %v\n", err)
		os.Exit(1)
	}

	var entries []SampleEntry
	if err := json.Unmarshal(indexData, &entries); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing index.json: %v\n", err)
		os.Exit(1)
	}

	var bundled []BundledSample

	for _, entry := range entries {
		metadataPath := filepath.Join(testSamplesDir, entry.Metadata)
		metadataData, err := os.ReadFile(metadataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot read metadata %s: %v\n", entry.Metadata, err)
			continue
		}

		var payload *string
		if entry.Payload != nil {
			payloadPath := filepath.Join(testSamplesDir, *entry.Payload)
			payloadData, err := os.ReadFile(payloadPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: cannot read payload %s: %v\n", *entry.Payload, err)
				continue
			}
			s := base64.StdEncoding.EncodeToString(payloadData)
			payload = &s
		}

		bundled = append(bundled, BundledSample{
			Sample:   entry.Sample,
			Protocol: entry.Protocol,
			Metadata: base64.StdEncoding.EncodeToString(metadataData),
			Payload:  payload,
		})
	}

	outPath := filepath.Join(testSamplesDir, "bundle.json")
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(bundled); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Bundled %d test samples to %s\n", len(bundled), outPath)
}
