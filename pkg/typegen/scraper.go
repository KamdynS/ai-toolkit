package typegen

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
)

// Maximum depth for recursion when extracting content
const maxDepth = 3

// ScrapeDocumentation fetches and extracts relevant content from an API documentation URL
func ScrapeDocumentation(docURL string, funcName string, verbose bool) (string, error) {
	// Validate URL
	_, err := url.ParseRequestURI(docURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	if verbose {
		log.Printf("Starting scraping of %s", docURL)
	}

	// Initialize a new collector
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
		colly.MaxDepth(2),
	)

	// Set timeout to prevent hanging
	c.SetRequestTimeout(30 * time.Second)

	// Store the extracted content
	var content strings.Builder

	// First, collect the title of the page
	c.OnHTML("title", func(e *colly.HTMLElement) {
		content.WriteString(fmt.Sprintf("# %s\n\n", e.Text))
	})

	// We'll collect content based on common documentation containers
	// This covers most documentation sites
	selectors := []string{
		"main", "article", ".documentation", ".docs", ".content",
		".main-content", ".api-docs", ".api-reference", "#documentation",
		".markdown-body", "[role='main']", ".api-content",
	}

	// If a specific function/method is specified, try to find it
	if funcName != "" {
		selectors = append([]string{
			fmt.Sprintf("[id='%s']", funcName),
			fmt.Sprintf("[id='%s-method']", funcName),
			fmt.Sprintf("[id='%s-function']", funcName),
			fmt.Sprintf("[id='method-%s']", funcName),
			fmt.Sprintf("[id='function-%s']", funcName),
			fmt.Sprintf("[name='%s']", funcName),
			fmt.Sprintf("[data-function='%s']", funcName),
			fmt.Sprintf("[data-method='%s']", funcName),
		}, selectors...)
	}

	// Collect content from these selectors
	for _, selector := range selectors {
		c.OnHTML(selector, func(e *colly.HTMLElement) {
			// If specific function requested, filter further
			if funcName != "" {
				// Check if the element contains the function name in text
				if strings.Contains(e.Text, funcName) {
					// Extract relevant section and append to content
					extractedText := extractRelevantSection(e.DOM, funcName, verbose)
					if extractedText != "" {
						content.WriteString(extractedText)
						content.WriteString("\n\n")
					}
				}
				return
			}

			// Otherwise, extract all content
			text := e.DOM.Text()
			// Clean up text a bit
			text = strings.TrimSpace(text)
			if text != "" {
				content.WriteString(text)
				content.WriteString("\n\n")
			}

			// Try to capture code blocks, tables, etc.
			e.DOM.Find("pre, code, table").Each(func(i int, s *goquery.Selection) {
				html, err := s.Html()
				if err == nil {
					content.WriteString("\n")
					content.WriteString(html)
					content.WriteString("\n")
				}
			})
		})
	}

	// Specifically look for code blocks which often contain type examples
	c.OnHTML("pre, code", func(e *colly.HTMLElement) {
		// If specific function is requested, only capture if related
		if funcName != "" && !strings.Contains(e.Text, funcName) {
			return
		}

		text := e.Text
		if text != "" {
			content.WriteString("\n```\n")
			content.WriteString(text)
			content.WriteString("\n```\n")
		}
	})

	// Visit the URL
	err = c.Visit(docURL)
	if err != nil {
		return "", fmt.Errorf("error visiting URL: %v", err)
	}

	// Wait for scraping to finish
	c.Wait()

	result := content.String()

	// If we didn't find anything specific but a function name was provided,
	// try a more aggressive approach with a second pass
	if funcName != "" && !strings.Contains(result, funcName) {
		if verbose {
			log.Printf("Function %s not found in first pass, trying second pass", funcName)
		}

		// Reset the content builder
		content.Reset()

		// New collector for second pass
		c2 := colly.NewCollector(
			colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
		)

		// Set timeout
		c2.SetRequestTimeout(30 * time.Second)

		// Look for any element that contains the function name
		c2.OnHTML("*", func(e *colly.HTMLElement) {
			// Check if this element or its children contain the function name
			if strings.Contains(e.Text, funcName) {
				// Extract only if it's a reasonable container element
				if isContentElement(e.Name) {
					extractedText := extractRelevantSection(e.DOM, funcName, verbose)
					if extractedText != "" {
						content.WriteString(extractedText)
						content.WriteString("\n\n")
					}
				}
			}
		})

		// Visit the URL
		err = c2.Visit(docURL)
		if err != nil {
			return "", fmt.Errorf("error in second visit to URL: %v", err)
		}

		// Wait for scraping to finish
		c2.Wait()

		// Update result
		result = content.String()
	}

	if result == "" {
		if funcName != "" {
			return "", fmt.Errorf("function or method '%s' not found in the documentation", funcName)
		}
		return "", errors.New("no content extracted from the documentation URL")
	}

	if verbose {
		contentPreview := result
		if len(contentPreview) > 200 {
			contentPreview = contentPreview[:200] + "..."
		}
		log.Printf("Extracted content preview: %s", contentPreview)
	}

	return result, nil
}

// extractRelevantSection extracts content relevant to a specific function
func extractRelevantSection(s *goquery.Selection, funcName string, verbose bool) string {
	var content strings.Builder

	// First check if this element itself is a good container for the function
	if isFunctionContainer(s, funcName) {
		// First try to extract just the section about this function
		functionSection := extractFunctionSection(s, funcName, 0)
		if functionSection != "" {
			return functionSection
		}

		// If that fails, extract the whole element
		html, err := s.Html()
		if err == nil {
			return html
		}
	}

	// Look for function container in children 
	s.Find("*").Each(func(i int, child *goquery.Selection) {
		if isFunctionContainer(child, funcName) {
			section := extractFunctionSection(child, funcName, 0)
			if section != "" {
				content.WriteString(section)
				content.WriteString("\n\n")
			}
		}
	})

	return content.String()
}

// isFunctionContainer checks if an element is likely to be a container for function documentation
func isFunctionContainer(s *goquery.Selection, funcName string) bool {
	// Check if the element contains the function name
	if !strings.Contains(s.Text(), funcName) {
		return false
	}

	// Check if it's an element that typically contains function documentation
	if isHeadingElement(s) {
		// If it's a heading, it might be the title of the function documentation
		return true
	}

	// Check if it has an ID or name attribute that suggests it's for this function
	id, _ := s.Attr("id")
	name, _ := s.Attr("name")
	dataFunction, _ := s.Attr("data-function")
	dataMethod, _ := s.Attr("data-method")

	relevantAttrs := []string{
		id, name, dataFunction, dataMethod,
		strings.ToLower(id), strings.ToLower(name),
		strings.ToLower(dataFunction), strings.ToLower(dataMethod),
	}

	for _, attr := range relevantAttrs {
		if attr == funcName || 
		   attr == fmt.Sprintf("%s-method", funcName) || 
		   attr == fmt.Sprintf("%s-function", funcName) ||
		   attr == fmt.Sprintf("method-%s", funcName) ||
		   attr == fmt.Sprintf("function-%s", funcName) {
			return true
		}
	}

	// Check if it's a section or div that's dedicated to this function
	nodeName := goquery.NodeName(s)
	if nodeName == "section" || nodeName == "div" || nodeName == "article" {
		// If we find a heading inside that contains exactly the function name
		var foundHeading bool
		s.Find("h1, h2, h3, h4, h5, h6").Each(func(i int, heading *goquery.Selection) {
			if strings.TrimSpace(heading.Text()) == funcName {
				foundHeading = true
				return
			}
		})
		return foundHeading
	}

	return false
}

// extractFunctionSection tries to extract just the section about a specific function
func extractFunctionSection(s *goquery.Selection, funcName string, depth int) string {
	if depth > maxDepth {
		return ""
	}

	var content strings.Builder

	// Check if this element is a heading that contains the function name
	if isHeadingElement(s) && strings.Contains(s.Text(), funcName) {
		// Get this heading and all content until the next heading of same or higher level
		headingLevel := getHeadingLevel(s)
		
		// Add the heading
		headingHTML, _ := s.Html()
		content.WriteString(fmt.Sprintf("<%s>%s</%s>\n", 
			goquery.NodeName(s), headingHTML, goquery.NodeName(s)))
		
		// Get all siblings until next heading of same or higher level
		var next *goquery.Selection = s.Next()
		for next.Length() > 0 {
			if isHeadingElement(next) {
				nextLevel := getHeadingLevel(next)
				if nextLevel <= headingLevel {
					break
				}
			}
			
			html, _ := next.Html()
			content.WriteString(html)
			content.WriteString("\n")
			
			next = next.Next()
		}
		
		return content.String()
	}

	// If this element contains code examples/snippets, include them
	s.Find("pre, code").Each(func(i int, code *goquery.Selection) {
		if strings.Contains(code.Text(), funcName) {
			html, _ := code.Html()
			content.WriteString(html)
			content.WriteString("\n")
		}
	})

	// If it's a list or table that might contain parameter/return info
	if goquery.NodeName(s) == "table" || goquery.NodeName(s) == "ul" || goquery.NodeName(s) == "ol" {
		if strings.Contains(s.Text(), funcName) {
			html, _ := s.Html()
			content.WriteString(html)
			content.WriteString("\n")
		}
	}

	return content.String()
}

// isHeadingElement checks if an element is a heading (h1-h6)
func isHeadingElement(s *goquery.Selection) bool {
	nodeName := goquery.NodeName(s)
	return nodeName == "h1" || nodeName == "h2" || nodeName == "h3" || 
	       nodeName == "h4" || nodeName == "h5" || nodeName == "h6"
}

// getHeadingLevel returns the level of a heading (1 for h1, 2 for h2, etc.)
func getHeadingLevel(s *goquery.Selection) int {
	nodeName := goquery.NodeName(s)
	switch nodeName {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	default:
		return 0
	}
}

// isContentElement checks if an element is typically used for content
func isContentElement(name string) bool {
	contentElements := []string{
		"div", "section", "article", "main", "p", "pre", "code",
		"table", "ul", "ol", "dl", "blockquote", "figure", "details",
		"summary", "aside", "header", "footer", "nav",
	}

	for _, elem := range contentElements {
		if name == elem {
			return true
		}
	}

	return false
}

// Helper function to render a node to HTML
func renderNode(n *html.Node) string {
	var buf strings.Builder
	html.Render(&buf, n)
	return buf.String()
}