package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Variant struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Script      string `json:"script"`
}

type Template struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Variants    []Variant `json:"variants"`
}

type ManifestVariant struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ScriptFile  string `json:"scriptFile"`
}

type Manifest struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Variants    []ManifestVariant `json:"variants"`
}

func main() {
	templatesDir := "templates"
	var templates []Template

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading templates dir: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		tmplDir := filepath.Join(templatesDir, entry.Name())
		manifestPath := filepath.Join(tmplDir, "manifest.json")

		manifestData, err := os.ReadFile(manifestPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", entry.Name(), err)
			continue
		}

		var manifest Manifest
		if err := json.Unmarshal(manifestData, &manifest); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: invalid manifest: %v\n", entry.Name(), err)
			continue
		}

		tmpl := Template{
			Name:        manifest.Name,
			Description: manifest.Description,
		}

		for _, mv := range manifest.Variants {
			scriptPath := filepath.Join(tmplDir, mv.ScriptFile)
			scriptData, err := os.ReadFile(scriptPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %s: cannot read %s: %v\n", entry.Name(), mv.ScriptFile, err)
				continue
			}

			tmpl.Variants = append(tmpl.Variants, Variant{
				Name:        mv.Name,
				Description: mv.Description,
				Script:      string(scriptData),
			})
		}

		templates = append(templates, tmpl)
	}

	outPath := filepath.Join(templatesDir, "bundle.json")
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(templates); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Bundled %d templates to %s\n", len(templates), outPath)
}
