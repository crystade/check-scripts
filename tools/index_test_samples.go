package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Sample struct {
	Sample   string  `json:"sample"`
	Protocol string  `json:"protocol"`
	Metadata string  `json:"metadata"`
	Payload  *string `json:"payload"`
}

func main() {
	srcDir := filepath.Join("test-samples", "src")
	var samples []Sample

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		relPath, err := filepath.Rel("test-samples", path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		dir := filepath.Dir(path)
		baseName := strings.TrimSuffix(filepath.Base(path), ".json")
		protocol := filepath.Base(dir)

		sample := Sample{
			Sample:   baseName,
			Protocol: protocol,
			Metadata: relPath,
		}

		txtPath := filepath.Join(dir, baseName+".txt")
		if _, err := os.Stat(txtPath); err == nil {
			txtRel, _ := filepath.Rel("test-samples", txtPath)
			txtRel = filepath.ToSlash(txtRel)
			sample.Payload = &txtRel
		}

		samples = append(samples, sample)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join("test-samples", "index.json")
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(samples); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Indexed %d samples to %s\n", len(samples), outPath)
}
