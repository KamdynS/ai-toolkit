package docgen

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/kamdyn/ai-toolkit/pkg/common"
)

// DocGenerator handles the generation of documentation
type DocGenerator struct {
	client *common.AIClient
}

// NewDocGenerator creates a new DocGenerator
func NewDocGenerator(client *common.AIClient) *DocGenerator {
	return &DocGenerator{
		client: client,
	}
}

// GenerateDocumentation generates documentation for code
func (g *DocGenerator) GenerateDocumentation(ctx context.Context, modelName string, temperature float32, code string, language string, style string, verbose bool) (string, error) {
	// Prepare the prompt based on the language and style
	prompt := buildPrompt(code, language, style)

	if verbose {
		log.Println("Sending request to Gemini API...")
	}

	// Generate content using the AI client
	result, err := g.client.Generate(ctx, prompt, modelName, temperature)
	if err != nil {
		return "", fmt.Errorf("error generating documentation: %v", err)
	}

	// Extract and format the output
	output := common.ExtractCode(result, getLanguageMarker(language))
	
	// For markdown documentation, we're good to go
	// For other styles, we need to format differently in the calling code
	
	return output, nil
}

// buildPrompt creates a prompt for the AI based on the code, language, and style
func buildPrompt(code string, language string, style string) string {
	var sb strings.Builder

	// Base prompt
	sb.WriteString(fmt.Sprintf("Generate high-quality documentation for the following %s code. ", getLanguageName(language)))
	
	// Style-specific instructions
	switch style {
	case "jsdoc", "tsdoc":
		sb.WriteString("Use JSDoc style with detailed descriptions, @param, @returns, @throws tags where appropriate. Include type information and examples where helpful. ")
	case "godoc":
		sb.WriteString("Follow Go's standard godoc convention. Start with a brief summary. Include example usage where appropriate. Document parameters and return values. ")
	case "docstring":
		sb.WriteString("Use Python docstring conventions following Google style. Include descriptions for the function/class, Args, Returns, Raises sections with type information. ")
	case "xml":
		sb.WriteString("Use XML documentation comments style (e.g., /// for C# or /** */ for Java). Include parameter descriptions, return value details, and exception information. ")
	case "markdown":
		sb.WriteString("Create Markdown documentation with proper headings, code blocks, parameter tables, and examples. Include detailed usage information. ")
	default:
		// Default style based on the language
		switch language {
		case "typescript", "ts", "javascript", "js":
			sb.WriteString("Use JSDoc style with detailed descriptions, @param, @returns, @throws tags. Include type information and examples where helpful. ")
		case "go", "golang":
			sb.WriteString("Follow Go's standard godoc convention. Start with a brief summary. Include example usage where appropriate. Document parameters and return values. ")
		case "python", "py":
			sb.WriteString("Use Python docstring conventions following Google style. Include descriptions for the function/class, Args, Returns, Raises sections with type information. ")
		case "java":
			sb.WriteString("Use JavaDoc style comments with @param, @return, and @throws tags. Include detailed descriptions for classes, methods, and fields. ")
		case "csharp", "cs":
			sb.WriteString("Use XML documentation comments (///) with <summary>, <param>, <returns>, and <exception> tags. ")
		case "rust", "rs":
			sb.WriteString("Use Rust's documentation syntax (///) following rustdoc conventions. Include examples in ```rust blocks. Document parameters, return values, and errors. ")
		default:
			sb.WriteString("Include detailed comments describing what the code does, parameters, return values, and examples where appropriate. ")
		}
	}

	// Additional instructions
	sb.WriteString("Ensure the documentation is comprehensive yet concise. Focus on explaining the purpose, usage, parameters, and return values. ")
	sb.WriteString("The documentation should be directly applicable to the code and ready to use without modifications. ")
	sb.WriteString("Maintain the original structure and formatting of the code, only adding documentation comments. ")
	sb.WriteString("Output both the documentation comments and the original code together as a complete documented file. ")

	// Add the code
	sb.WriteString("\n\nCODE TO DOCUMENT:\n```")
	sb.WriteString(language)
	sb.WriteString("\n")
	sb.WriteString(code)
	sb.WriteString("\n```")

	return sb.String()
}

// getLanguageName returns the full name of a language from its code
func getLanguageName(code string) string {
	switch code {
	case "typescript", "ts":
		return "TypeScript"
	case "javascript", "js":
		return "JavaScript"
	case "go", "golang":
		return "Go"
	case "python", "py":
		return "Python"
	case "rust", "rs":
		return "Rust"
	case "java":
		return "Java"
	case "csharp", "cs":
		return "C#"
	case "swift":
		return "Swift"
	case "kotlin", "kt":
		return "Kotlin"
	default:
		return code
	}
}

// getLanguageMarker returns the language marker used in code blocks
func getLanguageMarker(code string) string {
	switch code {
	case "typescript", "ts":
		return "typescript|ts"
	case "javascript", "js":
		return "javascript|js"
	case "go", "golang":
		return "go|golang"
	case "python", "py":
		return "python|py"
	case "rust", "rs":
		return "rust|rs"
	case "java":
		return "java"
	case "csharp", "cs":
		return "csharp|cs"
	case "swift":
		return "swift"
	case "kotlin", "kt":
		return "kotlin|kt"
	default:
		return code
	}
}