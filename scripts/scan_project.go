package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ProjectScanner struct {
	rootDir        string
	ignorePatterns []string
	structure      []string
	contents       []string
}

func NewProjectScanner(rootDir string) *ProjectScanner {
	return &ProjectScanner{
		rootDir: rootDir,
		ignorePatterns: []string{
			".git",
			"node_modules",
			".env",
			".idea",
			"vendor",
			"dist",
			"build",
			"mediatracker.log",
			"project_knowledge.md",
			"go.sum",
			"gonamer-cache.gob",
		},
		structure: make([]string, 0),
		contents:  make([]string, 0),
	}
}

func (ps *ProjectScanner) shouldIgnore(path string) bool {
	for _, pattern := range ps.ignorePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func (ps *ProjectScanner) addToStructure(path string, info fs.FileInfo, depth int) {
	indent := strings.Repeat("  ", depth)
	if info.IsDir() {
		ps.structure = append(ps.structure, fmt.Sprintf("%s- 📁 %s", indent, info.Name()))
	} else {
		ps.structure = append(ps.structure, fmt.Sprintf("%s- 📄 %s", indent, info.Name()))
	}
}

func (ps *ProjectScanner) addToContents(path string, relPath string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	ext := filepath.Ext(path)
	if ext != "" {
		ext = ext[1:] // Remove the dot
	}

	fileContent := fmt.Sprintf("\n# 📄 %s\n```%s\n%s\n```\n",
		relPath, ext, string(content))
	ps.contents = append(ps.contents, fileContent)
	return nil
}

func (ps *ProjectScanner) scan() error {
	return filepath.Walk(ps.rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Skip ignored patterns
		if ps.shouldIgnore(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(ps.rootDir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path for %s: %w", path, err)
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Calculate depth for indentation
		depth := len(strings.Split(relPath, string(os.PathSeparator))) - 1

		// Add to structure
		ps.addToStructure(path, info, depth)

		// If it's a file, add its contents
		if !info.IsDir() {
			if err := ps.addToContents(path, relPath); err != nil {
				fmt.Printf("Warning: %v\n", err)
			}
		}

		return nil
	})
}

func (ps *ProjectScanner) generateOutput() string {
	return fmt.Sprintf("# Project Structure\n\n%s\n\n# Files Content\n%s",
		strings.Join(ps.structure, "\n"),
		strings.Join(ps.contents, "\n"))
}

func main() {
	outputFile := flag.String("o", "project_knowledge.md", "Output file path")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: Please provide the root directory path")
		fmt.Println("Usage: scan-project <root-dir> [-o output-file]")
		os.Exit(1)
	}

	rootDir := args[0]
	scanner := NewProjectScanner(rootDir)

	if err := scanner.scan(); err != nil {
		fmt.Printf("Error scanning project: %v\n", err)
		os.Exit(1)
	}

	output := scanner.generateOutput()
	if err := os.WriteFile(*outputFile, []byte(output), 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated project structure in %s\n", *outputFile)
}
