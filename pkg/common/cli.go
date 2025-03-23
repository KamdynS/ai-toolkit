package common

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

// ToolConfig represents common configuration for AI tools
type ToolConfig struct {
	APIKey      string
	Model       string
	Temperature float32
	Timeout     int
	Verbose     bool
	OutputFile  string
}

// CommonFlags returns common CLI flags used across tools
func CommonFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "api-key",
			Aliases: []string{"k"},
			Usage:   "Gemini API key (can also be set with GEMINI_API_KEY environment variable)",
			EnvVars: []string{"GEMINI_API_KEY"},
		},
		&cli.StringFlag{
			Name:    "model",
			Usage:   "Gemini model to use",
			Value:   GetEnvOrDefault("DEFAULT_MODEL", DefaultAIModel),
			EnvVars: []string{"DEFAULT_MODEL"},
		},
		&cli.Float64Flag{
			Name:    "temp",
			Usage:   "Temperature for generation (0.0-1.0)",
			Value:   GetEnvOrDefaultFloat("DEFAULT_TEMPERATURE", 0.2),
			EnvVars: []string{"DEFAULT_TEMPERATURE"},
		},
		&cli.IntFlag{
			Name:    "timeout",
			Usage:   "Timeout in seconds",
			Value:   GetEnvOrDefaultInt("DEFAULT_TIMEOUT", 120),
			EnvVars: []string{"DEFAULT_TIMEOUT"},
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Usage:   "Enable verbose logging",
			Value:   GetEnvOrDefaultBool("DEFAULT_VERBOSE", false),
			EnvVars: []string{"DEFAULT_VERBOSE"},
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "Output file path",
		},
	}
}

// PrepareLogger configures logging based on verbosity
func PrepareLogger(toolName string, verbose bool) {
	// Configure logging
	log.SetPrefix(fmt.Sprintf("[%s] ", toolName))
	
	if verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(0)
	}
}

// ValidateAPIKey checks if an API key is provided
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key is required. Provide it using --api-key flag or GEMINI_API_KEY environment variable")
	}
	return nil
}

// ExtractCommonConfig extracts common configuration from CLI context
func ExtractCommonConfig(c *cli.Context) ToolConfig {
	return ToolConfig{
		APIKey:      c.String("api-key"),
		Model:       c.String("model"),
		Temperature: float32(c.Float64("temp")),
		Timeout:     c.Int("timeout"),
		Verbose:     c.Bool("verbose"),
		OutputFile:  c.String("output"),
	}
}

// WriteOutput writes content to a file or stdout
func WriteOutput(content, outputFile string, verbose bool) error {
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("error writing to output file: %v", err)
		}
		
		if verbose {
			log.Printf("Output written to %s", outputFile)
		} else {
			fmt.Printf("Output written to %s\n", outputFile)
		}
	} else {
		// Write to stdout
		fmt.Println(content)
	}
	
	return nil
}