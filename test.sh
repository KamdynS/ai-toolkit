#!/bin/bash

# Set strict mode
set -e
set -o pipefail

# Define colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=====================================${NC}"
echo -e "${GREEN}AI Toolkit Test Script${NC}"
echo -e "${BLUE}=====================================${NC}"

# Load .env file if it exists
if [ -f .env ]; then
    echo -e "${BLUE}Loading API key from .env file...${NC}"
    # Handle potential Windows line endings by removing any carriage returns
    export GEMINI_API_KEY=$(grep "GEMINI_API_KEY" .env | sed 's/^GEMINI_API_KEY=//' | tr -d '\r')
    echo -e "${GREEN}Loaded API key: ${GEMINI_API_KEY:0:5}... (truncated for security)${NC}"
fi

# Check if the API key is provided
if [ -z "$GEMINI_API_KEY" ]; then
    echo -e "${RED}Error: GEMINI_API_KEY environment variable is not set${NC}"
    echo -e "${RED}Please set your API key in the .env file or with: export GEMINI_API_KEY=your-api-key-here${NC}"
    exit 1
fi

# Make sure the binaries exist
if [ ! -f bin/ai-tools ] || [ ! -f bin/ai-tools-typegen ] || [ ! -f bin/ai-tools-docgen ]; then
    echo -e "${BLUE}Building AI Toolkit...${NC}"
    ./install.sh
fi

# Test variables
TEST_URL="https://docs.stripe.com/api/charges"
TYPES_OUTPUT="test-output/stripe-charges.ts"
DOC_INPUT="test-input/sample.go"
DOC_OUTPUT="test-output/sample.doc.go"

# Create test directories
mkdir -p test-output
mkdir -p test-input

# Create sample Go file for documentation testing
cat > test-input/sample.go << 'EOF'
package main

import "fmt"

// A simple function that adds two numbers
func add(a, b int) int {
	return a + b
}

func main() {
	result := add(5, 7)
	fmt.Printf("The result is: %d\n", result)
}
EOF

echo -e "\n${BLUE}Testing TypeGen with Stripe API docs...${NC}"
echo -e "${BLUE}Generating TypeScript types for the Stripe Charges API...${NC}"
./bin/ai-tools-typegen --url="$TEST_URL" --output="$TYPES_OUTPUT" --verbose

# Check if the output file was created
if [ -f "$TYPES_OUTPUT" ]; then
    echo -e "\n${GREEN}TypeGen test successful!${NC}"
    echo -e "${BLUE}First 10 lines of the generated file:${NC}"
    head -n 10 "$TYPES_OUTPUT"
else
    echo -e "\n${RED}TypeGen test failed: Output file was not created${NC}"
    exit 1
fi

echo -e "\n${BLUE}Testing DocGen with a sample Go file...${NC}"
echo -e "${BLUE}Generating documentation for a sample Go file...${NC}"
./bin/ai-tools-docgen --file="$DOC_INPUT" --output="$DOC_OUTPUT" --verbose

# Check if the output file was created
if [ -f "$DOC_OUTPUT" ]; then
    echo -e "\n${GREEN}DocGen test successful!${NC}"
    echo -e "${BLUE}First 10 lines of the generated file:${NC}"
    head -n 10 "$DOC_OUTPUT"
else
    echo -e "\n${RED}DocGen test failed: Output file was not created${NC}"
    exit 1
fi

echo -e "\n${GREEN}All tests passed!${NC}"
echo -e "${BLUE}=====================================${NC}"
echo -e "${GREEN}AI Toolkit is working correctly.${NC}"
echo -e "${BLUE}=====================================${NC}"