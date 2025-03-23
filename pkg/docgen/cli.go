package docgen

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kamdyn/ai-toolkit/pkg/common"
	"github.com/urfave/cli/v2"
)

// GetDocGenCommand returns the CLI command for docgen
func GetDocGenCommand() *cli.Command {
	return &cli.Command{
		Name:  "docgen",
		Usage: "Generate documentation for code",
		Flags: append(common.CommonFlags(),
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Source code file to document",
			},
			&cli.StringFlag{
				Name:    "dir",
				Aliases: []string{"d"},
				Usage:   "Directory to generate documentation for (generates PROJECT.md)",
			},
			&cli.StringFlag{
				Name:    "lang",
				Aliases: []string{"l"},
				Usage:   "Language of the source code (auto-detected from file extension if not specified)",
			},
			&cli.StringFlag{
				Name:    "style",
				Aliases: []string{"s"},
				Usage:   "Documentation style (jsdoc, godoc, docstring, xml, markdown)",
			},
			&cli.StringFlag{
				Name:    "title",
				Aliases: []string{"t"},
				Usage:   "Project title for documentation (only used with --dir)",
				Value:   "Project Documentation",
			},
		),
		Before: func(c *cli.Context) error {
			// Validate API key
			if err := common.ValidateAPIKey(c.String("api-key")); err != nil {
				return err
			}

			// Check that at least one of --file or --dir is provided
			filePath := c.String("file")
			dirPath := c.String("dir")
			if filePath == "" && dirPath == "" {
				return fmt.Errorf("either --file or --dir must be provided")
			}

			// If file is provided, validate it exists
			if filePath != "" {
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					return fmt.Errorf("file does not exist: %s", filePath)
				}
			}

			// If directory is provided, validate it exists
			if dirPath != "" {
				if _, err := os.Stat(dirPath); os.IsNotExist(err) {
					return fmt.Errorf("directory does not exist: %s", dirPath)
				}
			}

			return nil
		},
		Action: func(c *cli.Context) error {
			return runDocGen(c)
		},
	}
}

// runDocGen runs the documentation generator
func runDocGen(c *cli.Context) error {
	// Extract configuration
	config := common.ExtractCommonConfig(c)
	
	filePath := c.String("file")
	dirPath := c.String("dir")
	language := c.String("lang")
	style := c.String("style")
	projectTitle := c.String("title")
	
	// Configure logging based on verbose flag
	common.PrepareLogger("DocGen", config.Verbose)

	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	// Create AI client
	aiClient, err := common.NewAIClient(ctx, config.APIKey)
	if err != nil {
		return fmt.Errorf("error creating AI client: %v", err)
	}
	defer aiClient.Close()

	// Create generator
	generator := NewDocGenerator(aiClient)

	// If directory is provided, generate project documentation
	if dirPath != "" {
		return generateProjectDocumentation(ctx, generator, dirPath, projectTitle, config)
	}

	// Otherwise, generate documentation for a single file
	return generateFileDocumentation(ctx, generator, filePath, language, style, config)
}

// generateFileDocumentation generates documentation for a single file
func generateFileDocumentation(ctx context.Context, generator *DocGenerator, filePath, language, style string, config common.ToolConfig) error {
	// If language not provided, try to detect from file extension
	if language == "" {
		language = detectLanguage(filePath)
		if config.Verbose {
			log.Printf("Auto-detected language: %s", language)
		}
	}

	// If no output file specified, create one in the docs directory
	if config.OutputFile == "" {
		// Create docs directory if it doesn't exist
		docsDir := "docs"
		if err := os.MkdirAll(docsDir, 0755); err != nil {
			return fmt.Errorf("error creating docs directory: %v", err)
		}
		
		// For single file documentation, use a standardized naming convention
		relPath := filePath
		// Convert path separators to underscores for flat structure
		fileName := strings.ReplaceAll(relPath, "/", "_")
		
		// Always use .md extension for documentation files in the docs directory
		config.OutputFile = filepath.Join(docsDir, fileName+".md")
	}

	if config.Verbose {
		log.Printf("Generating documentation for file: %s", filePath)
		log.Printf("Language: %s", language)
		if style != "" {
			log.Printf("Documentation style: %s", style)
		}
		log.Printf("Output file: %s", config.OutputFile)
	}

	// Read the source code
	codeBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}
	code := string(codeBytes)

	// Generate documentation
	if config.Verbose {
		log.Printf("Generating documentation using model %s...", config.Model)
	}
	
	// If we're generating markdown-style documentation, always use "markdown" style
	if strings.HasSuffix(config.OutputFile, ".md") {
		style = "markdown"
	}
	
	documentation, err := generator.GenerateDocumentation(ctx, config.Model, config.Temperature, code, language, style, config.Verbose)
	if err != nil {
		return fmt.Errorf("error generating documentation: %v", err)
	}

	// If we're writing to markdown file, ensure it has the right format
	if strings.HasSuffix(config.OutputFile, ".md") {
		// Add file title if not already present
		if !strings.HasPrefix(documentation, "# ") {
			fileName := filepath.Base(filePath)
			documentation = fmt.Sprintf("# Documentation for %s\n\n%s", fileName, documentation)
		}
	}

	if config.Verbose {
		log.Println("Documentation successfully generated")
	}

	// Write to output file
	return common.WriteOutput(documentation, config.OutputFile, config.Verbose)
}

// generateProjectDocumentation generates documentation for a project directory
func generateProjectDocumentation(ctx context.Context, generator *DocGenerator, dirPath, projectTitle string, config common.ToolConfig) error {
	if config.Verbose {
		log.Printf("Generating project documentation for directory: %s", dirPath)
	}

	// If no output file specified, use PROJECT.md in the root directory
	if config.OutputFile == "" {
		config.OutputFile = filepath.Join(dirPath, "PROJECT.md")
	}

	// Create docs directory if it doesn't exist
	docsDir := filepath.Join(dirPath, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("error creating docs directory: %v", err)
	}

	// Find all source code files in the directory
	var codeFiles []string
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the docs directory itself
		rel, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, "docs/") || strings.HasPrefix(rel, ".git/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process regular files
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			// Only include source code files
			if isSourceCodeFile(ext) {
				codeFiles = append(codeFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %v", err)
	}

	if len(codeFiles) == 0 {
		return fmt.Errorf("no source code files found in directory: %s", dirPath)
	}

	if config.Verbose {
		log.Printf("Found %d source code files", len(codeFiles))
	}

	// Generate documentation for each file
	fileInfos := make([]FileDocInfo, 0, len(codeFiles))
	for _, file := range codeFiles {
		relPath, err := filepath.Rel(dirPath, file)
		if err != nil {
			return fmt.Errorf("error getting relative path: %v", err)
		}

		language := detectLanguage(file)
		outputPath := filepath.Join(docsDir, strings.ReplaceAll(relPath, "/", "_")+".md")

		if config.Verbose {
			log.Printf("Generating documentation for %s (%s)", relPath, language)
		}

		// Read the source code
		codeBytes, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Warning: error reading file %s: %v", file, err)
			continue
		}
		code := string(codeBytes)

		// Generate documentation 
		documentation, err := generator.GenerateDocumentation(ctx, config.Model, config.Temperature, code, language, "markdown", config.Verbose)
		if err != nil {
			log.Printf("Warning: error generating documentation for %s: %v", file, err)
			continue
		}

		// Write the documentation to a file in the docs directory
		err = os.WriteFile(outputPath, []byte(documentation), 0644)
		if err != nil {
			log.Printf("Warning: error writing documentation for %s: %v", file, err)
			continue
		}

		// Store file info for the combined documentation
		fileInfos = append(fileInfos, FileDocInfo{
			RelativePath: relPath,
			Language:     language,
			DocPath:      outputPath,
		})

		if config.Verbose {
			log.Printf("Documentation for %s written to %s", relPath, outputPath)
		}
	}

	// Create a combined documentation file
	if config.Verbose {
		log.Printf("Creating combined documentation file: %s", config.OutputFile)
	}

	err = createCombinedDocumentation(fileInfos, projectTitle, config.OutputFile)
	if err != nil {
		return fmt.Errorf("error creating combined documentation: %v", err)
	}

	if config.Verbose {
		log.Printf("Project documentation successfully written to %s", config.OutputFile)
	} else {
		fmt.Printf("Project documentation successfully written to %s\n", config.OutputFile)
	}

	return nil
}

// FileDocInfo holds information about a documented file
type FileDocInfo struct {
	RelativePath string
	Language     string
	DocPath      string
}

// createCombinedDocumentation creates a combined documentation file from individual file documentations
func createCombinedDocumentation(fileInfos []FileDocInfo, title string, outputPath string) error {
	var sb strings.Builder

	// Write the header
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	sb.WriteString("This document provides comprehensive documentation for the project.\n\n")

	// Write the table of contents
	sb.WriteString("## Table of Contents\n\n")
	
	// Add project overview
	sb.WriteString("1. [Project Overview](#project-overview)\n")
	
	// Group files by directory
	dirMap := make(map[string][]FileDocInfo)
	for _, info := range fileInfos {
		dir := filepath.Dir(info.RelativePath)
		if dir == "." {
			dir = "Root"
		}
		dirMap[dir] = append(dirMap[dir], info)
	}

	// Add TOC entries for each directory
	i := 2
	for dir := range dirMap {
		anchor := strings.ReplaceAll(strings.ToLower(dir), " ", "-")
		anchor = strings.ReplaceAll(anchor, "/", "")
		sb.WriteString(fmt.Sprintf("%d. [%s](#%s)\n", i, dir, anchor))
		i++
	}
	sb.WriteString("\n")

	// Add project overview
	sb.WriteString("## Project Overview\n\n")
	sb.WriteString("The project contains the following key components:\n\n")
	
	for dir, files := range dirMap {
		sb.WriteString(fmt.Sprintf("- **%s**: ", dir))
		fileDescs := make([]string, 0, len(files))
		for _, file := range files {
			fileDescs = append(fileDescs, filepath.Base(file.RelativePath))
		}
		sb.WriteString(strings.Join(fileDescs, ", "))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Add documentation for each directory
	for dir, files := range dirMap {
		dirHeader := dir
		sb.WriteString(fmt.Sprintf("## %s\n\n", dirHeader))
		
		// Add documentation for each file in the directory
		for _, file := range files {
			// Read the documentation file
			docContent, err := os.ReadFile(file.DocPath)
			if err != nil {
				return fmt.Errorf("error reading doc file %s: %v", file.DocPath, err)
			}
			
			// Add file header
			sb.WriteString(fmt.Sprintf("### %s\n\n", filepath.Base(file.RelativePath)))
			
			// Add documentation content
			sb.WriteString(string(docContent))
			sb.WriteString("\n\n")
		}
	}

	// Write the combined documentation to the output file
	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

// isSourceCodeFile checks if a file is a source code file based on its extension
func isSourceCodeFile(ext string) bool {
	ext = strings.ToLower(ext)
	sourceExtensions := map[string]bool{
		".go":   true,
		".js":   true,
		".ts":   true,
		".tsx":  true,
		".py":   true,
		".java": true,
		".c":    true,
		".cpp":  true,
		".cs":   true,
		".rs":   true,
		".rb":   true,
		".php":  true,
		".swift": true,
		".kt":   true,
		".sh":   true,
	}
	
	return sourceExtensions[ext]
}

// detectLanguage attempts to determine language from file extension
func detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	switch ext {
	case ".js":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".cs":
		return "csharp"
	case ".swift":
		return "swift"
	case ".kt":
		return "kotlin"
	case ".rb":
		return "ruby"
	case ".c", ".cpp", ".cc", ".h", ".hpp":
		return "cpp"
	case ".php":
		return "php"
	case ".sh":
		return "bash"
	default:
		return "text"
	}
}