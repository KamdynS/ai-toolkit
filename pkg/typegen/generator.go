package typegen

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kamdyn/ai-toolkit/pkg/common"
)

// TypeGenerator handles the generation of type definitions
type TypeGenerator struct {
	client *common.AIClient
}

// NewTypeGenerator creates a new TypeGenerator
func NewTypeGenerator(client *common.AIClient) *TypeGenerator {
	return &TypeGenerator{
		client: client,
	}
}

// GenerateTypeDefinitions generates type definitions from documentation content
func (g *TypeGenerator) GenerateTypeDefinitions(ctx context.Context, modelName string, temperature float32, docContent string, language string, funcName string, verbose bool) (string, error) {
	// Prepare the prompt based on the language
	prompt := buildPrompt(docContent, language, funcName)

	// Generate content using the AI client
	result, err := g.client.Generate(ctx, prompt, modelName, temperature)
	if err != nil {
		return "", fmt.Errorf("error generating type definitions: %v", err)
	}

	// Extract and format the output
	output := extractOutput(result, language)

	return output, nil
}

// buildPrompt creates a prompt for the AI based on the documentation and target language
func buildPrompt(docContent string, language string, funcName string) string {
	var sb strings.Builder

	// Base prompt
	sb.WriteString("Based on the following API documentation, ")
	
	if funcName != "" {
		sb.WriteString(fmt.Sprintf("generate accurate and complete type definitions for the function or method named '%s' ", funcName))
	} else {
		sb.WriteString("generate accurate and complete type definitions ")
	}
	
	sb.WriteString(fmt.Sprintf("in %s. ", getLanguageName(language)))

	// Language-specific instructions
	switch language {
	case "typescript", "ts":
		sb.WriteString("Include proper TypeScript interfaces, types, enums, generics, and all necessary types including parameters, request/response objects, and return types. Use strict typing (avoid 'any' when possible). Add JSDoc comments for all types. ")
	case "go", "golang":
		sb.WriteString("Create Go structs with proper field types and appropriate struct tags (json, xml, etc. as needed). Include interfaces, type aliases, and constants where appropriate. Add godoc style comments. ")
	case "python", "py":
		sb.WriteString("Use modern Python type annotations (typing module). Include type hints for function parameters, return types, class attributes, etc. Use dataclasses or Pydantic models where appropriate. Add docstrings for all types. ")
	case "rust", "rs":
		sb.WriteString("Create Rust structs and enums with proper field types. Include trait implementations, derive macros, and proper documentation comments. Use appropriate Serde annotations for serialization if needed. ")
	case "java":
		sb.WriteString("Create Java classes with proper field types, getters, setters and constructors. Include interfaces, enums, and generics where appropriate. Add Javadoc comments. Use appropriate annotations (e.g., Jackson annotations for JSON processing). ")
	case "csharp", "cs":
		sb.WriteString("Create C# classes with proper field types, properties, and constructors. Include interfaces, enums, and generics where appropriate. Add XML documentation comments. Use appropriate attributes (e.g., JsonProperty for JSON processing). ")
	case "swift":
		sb.WriteString("Create Swift structs/classes with proper field types and codable conformance where appropriate. Include protocols, enums, and optionals where needed. Add documentation comments. ")
	case "kotlin", "kt":
		sb.WriteString("Create Kotlin data classes with proper field types. Include interfaces, sealed classes, and nullable types where appropriate. Add KDoc comments. Use appropriate annotations (e.g., Serializable, JsonProperty). ")
	}

	// Additional instructions for all languages
	sb.WriteString("Ensure all types accurately represent the API's data structures, parameter types, and return values. Do not include implementation logic, only type definitions. ")
	
	if funcName != "" {
		sb.WriteString(fmt.Sprintf("Focus specifically on the '%s' function/method and its associated types. ", funcName))
	}

	sb.WriteString("Include all necessary imports/includes at the top of the file. ")
	sb.WriteString("Output only code, no additional explanation. ")

	// Add the documentation content
	sb.WriteString("\n\nAPI DOCUMENTATION:\n")
	sb.WriteString(docContent)

	return sb.String()
}

// extractOutput processes the API response and returns formatted output
func extractOutput(text string, language string) string {
	var result strings.Builder

	// Try to extract code blocks using language-specific markers
	codeBlockPattern := fmt.Sprintf("```(?:%s)?([\\s\\S]*?)```", getLanguageMarker(language))
	re := regexp.MustCompile(codeBlockPattern)
	matches := re.FindAllStringSubmatch(text, -1)

	if len(matches) > 0 {
		// Extract code from inside the code blocks
		for _, match := range matches {
			if len(match) > 1 {
				code := strings.TrimSpace(match[1])
				result.WriteString(code)
				result.WriteString("\n\n")
			}
		}
	} else {
		// If no code blocks found, try to clean up the text by removing markdown-like content
		// Remove headers
		text = regexp.MustCompile(`(?m)^#+ .*$`).ReplaceAllString(text, "")
		// Remove bullet points
		text = regexp.MustCompile(`(?m)^[*-] `).ReplaceAllString(text, "")

		result.WriteString(text)
	}

	return strings.TrimSpace(result.String())
}

// getLanguageName returns the full name of a language from its code
func getLanguageName(code string) string {
	switch code {
	case "typescript", "ts":
		return "TypeScript"
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