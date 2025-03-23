package typegen

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kamdyn/ai-toolkit/pkg/common"
	"github.com/urfave/cli/v2"
)

// GetTypeGenCommand returns the CLI command for the type generator
func GetTypeGenCommand() *cli.Command {
	return &cli.Command{
		Name:  "typegen",
		Usage: "Generate type definitions from API documentation",
		Flags: append(common.CommonFlags(),
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Usage:    "Documentation URL to scrape",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "func",
				Aliases: []string{"f"},
				Usage:   "Specific function or method to get types for",
			},
			&cli.StringFlag{
				Name:    "lang",
				Aliases: []string{"l"},
				Usage:   "Output language (typescript, go, python, rust, java, csharp, swift, kotlin)",
				Value:   common.GetEnvOrDefault("DEFAULT_LANG", "typescript"),
				EnvVars: []string{"DEFAULT_LANG"},
			},
		),
		Before: func(c *cli.Context) error {
			// Validate API key
			if err := common.ValidateAPIKey(c.String("api-key")); err != nil {
				return err
			}

			// Validate URL
			url := c.String("url")
			if url == "" {
				return fmt.Errorf("URL is required")
			}

			return nil
		},
		Action: func(c *cli.Context) error {
			return runTypeGen(c)
		},
	}
}

// runTypeGen runs the type generator
func runTypeGen(c *cli.Context) error {
	// Extract configuration
	config := common.ExtractCommonConfig(c)
	
	docURL := c.String("url")
	funcName := c.String("func")
	language := normalizeLanguage(c.String("lang"))
	
	// Configure logging based on verbose flag
	common.PrepareLogger("TypeGen", config.Verbose)

	// If no output file specified, use default based on language
	if config.OutputFile == "" {
		ext := getFileExtension(language)
		config.OutputFile = fmt.Sprintf("types%s", ext)
	}

	if config.Verbose {
		log.Printf("Starting type generation from URL: %s", docURL)
		if funcName != "" {
			log.Printf("Focusing on function/method: %s", funcName)
		}
		log.Printf("Output language: %s", language)
		log.Printf("Output file: %s", config.OutputFile)
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	// 1. Scrape the documentation URL
	if config.Verbose {
		log.Printf("Scraping documentation from %s...", docURL)
	}
	
	docContent, err := ScrapeDocumentation(docURL, funcName, config.Verbose)
	if err != nil {
		return fmt.Errorf("error scraping documentation: %v", err)
	}

	if docContent == "" {
		return fmt.Errorf("no content extracted from the documentation URL")
	}

	if config.Verbose {
		preview := docContent
		if len(preview) > 200 {
			preview = preview[:200] + "... [truncated]"
		}
		log.Printf("Documentation successfully scraped (%d bytes)", len(docContent))
		log.Printf("Content preview: %s", strings.ReplaceAll(preview, "\n", " "))
	}

	// 2. Create AI client
	aiClient, err := common.NewAIClient(ctx, config.APIKey)
	if err != nil {
		return fmt.Errorf("error creating AI client: %v", err)
	}
	defer aiClient.Close()

	// 3. Create generator
	generator := NewTypeGenerator(aiClient)

	// 4. Generate type definitions
	if config.Verbose {
		log.Printf("Generating type definitions using model %s...", config.Model)
	}
	
	typeDefinitions, err := generator.GenerateTypeDefinitions(ctx, config.Model, config.Temperature, docContent, language, funcName, config.Verbose)
	if err != nil {
		return fmt.Errorf("error generating type definitions: %v", err)
	}

	if config.Verbose {
		log.Println("Type definitions successfully generated")
	}

	// 5. Write to output file
	err = os.WriteFile(config.OutputFile, []byte(typeDefinitions), 0644)
	if err != nil {
		return fmt.Errorf("error writing to output file: %v", err)
	}

	if config.Verbose {
		log.Printf("Types successfully written to %s", config.OutputFile)
	} else {
		fmt.Printf("Types successfully written to %s\n", config.OutputFile)
	}

	return nil
}

// normalizeLanguage normalizes language names to standard format
func normalizeLanguage(lang string) string {
	// Convert to lowercase
	lang = strings.ToLower(lang)

	// Map language variations to standard names
	switch lang {
	case "ts", "typescript":
		return "typescript"
	case "go", "golang":
		return "go"
	case "py", "python":
		return "python"
	case "rs", "rust":
		return "rust"
	case "js", "javascript":
		return "javascript"
	case "cs", "csharp", "c#":
		return "csharp"
	case "kt", "kotlin":
		return "kotlin"
	default:
		return lang
	}
}

// getFileExtension returns the appropriate file extension for a language
func getFileExtension(language string) string {
	switch language {
	case "typescript", "ts":
		return ".ts"
	case "javascript", "js":
		return ".js"
	case "go", "golang":
		return ".go"
	case "python", "py":
		return ".py"
	case "rust", "rs":
		return ".rs"
	case "java":
		return ".java"
	case "csharp", "cs":
		return ".cs"
	case "swift":
		return ".swift"
	case "kotlin", "kt":
		return ".kt"
	default:
		return ".txt"
	}
}