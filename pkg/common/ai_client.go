package common

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// AIClient provides a unified interface for working with Gemini AI
type AIClient struct {
	client *genai.Client
}

// NewAIClient creates a new AIClient with the given API key
func NewAIClient(ctx context.Context, apiKey string) (*AIClient, error) {
	// Clean any carriage returns from the API key
	apiKey = strings.ReplaceAll(apiKey, "\r", "")
	
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creating Gemini client: %v", err)
	}
	
	return &AIClient{
		client: client,
	}, nil
}

// Close closes the underlying AI client
func (c *AIClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Generate generates content based on the given prompt
func (c *AIClient) Generate(ctx context.Context, prompt string, model string, temperature float32) (string, error) {
	// Use the specified model
	genModel := c.client.GenerativeModel(model)
	
	// Set generation parameters
	genModel.SetTemperature(temperature)

	// Generate content
	resp, err := genModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error generating content: %v", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", fmt.Errorf("no response generated from the AI")
	}

	// Extract text from the response
	var result strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		result.WriteString(fmt.Sprintf("%v", part))
	}

	return result.String(), nil
}

// ExtractCode extracts code blocks from text, optionally filtering by language
func ExtractCode(text, language string) string {
	var result strings.Builder
	
	// Try to extract code blocks using language-specific markers
	codeBlockPattern := "```(?:.*?)([\\s\\S]*?)```"
	if language != "" {
		// Try a language-specific pattern first
		langPattern := fmt.Sprintf("```(?:%s)?([\\s\\S]*?)```", language)
		matches := extractWithRegex(text, langPattern)
		
		if len(matches) > 0 {
			for _, match := range matches {
				result.WriteString(match)
				result.WriteString("\n\n")
			}
			return strings.TrimSpace(result.String())
		}
	}
	
	// Fall back to any code blocks
	matches := extractWithRegex(text, codeBlockPattern)
	if len(matches) > 0 {
		for _, match := range matches {
			result.WriteString(match)
			result.WriteString("\n\n")
		}
		return strings.TrimSpace(result.String())
	}
	
	// If no code blocks found, try to clean up the text by removing markdown-like content
	cleanText := text
	
	// Remove headers
	cleanText = regexp.MustCompile(`(?m)^#+ .*$`).ReplaceAllString(cleanText, "")
	
	// Remove bullet points
	cleanText = regexp.MustCompile(`(?m)^[*-] `).ReplaceAllString(cleanText, "")
	
	return strings.TrimSpace(cleanText)
}

// extractWithRegex extracts text matching the given pattern
func extractWithRegex(text, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(text, -1)
	
	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, strings.TrimSpace(match[1]))
		}
	}
	
	return results
}