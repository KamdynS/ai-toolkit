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
echo -e "${GREEN}AI Toolkit Installer${NC}"
echo -e "${BLUE}=====================================${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed. Please install Go 1.21 or later.${NC}"
    exit 1
fi

# Get Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)

# Check Go version
if [[ "$GO_MAJOR" -lt 1 ]] || ([[ "$GO_MAJOR" -eq 1 ]] && [[ "$GO_MINOR" -lt 21 ]]); then
    echo -e "${RED}Error: Go version 1.21 or later is required. You have Go $GO_VERSION.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Go version $GO_VERSION found${NC}"

# Check for Gemini API key
if [ -f .env ]; then
    echo -e "${GREEN}✓ .env file found${NC}"
else
    echo -e "${BLUE}Creating .env file...${NC}"
    cp .env.example .env
    echo -e "${GREEN}✓ Created .env file${NC}"
    echo -e "${RED}⚠ Please edit .env and add your Gemini API key${NC}"
fi

# Install dependencies
echo -e "${BLUE}Installing dependencies...${NC}"
go mod tidy
echo -e "${GREEN}✓ Dependencies installed${NC}"

# Create bin directory if it doesn't exist
mkdir -p bin

# Build main command
echo -e "${BLUE}Building ai-tools...${NC}"
go build -o bin/ai-tools main.go
echo -e "${GREEN}✓ Built ai-tools${NC}"

# Build individual commands
echo -e "${BLUE}Building ai-tools-typegen...${NC}"
go build -o bin/ai-tools-typegen cmd/typegen/main.go
echo -e "${GREEN}✓ Built ai-tools-typegen${NC}"

echo -e "${BLUE}Building ai-tools-docgen...${NC}"
go build -o bin/ai-tools-docgen cmd/docgen/main.go
echo -e "${GREEN}✓ Built ai-tools-docgen${NC}"

# Make binaries executable
chmod +x bin/*

# Create installation instructions
echo -e "${BLUE}=====================================${NC}"
echo -e "${GREEN}Installation complete!${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""
echo -e "To add AI Toolkit to your PATH, you can run:"
echo -e "${BLUE}export PATH=\"$PWD/bin:\$PATH\"${NC}"
echo ""
echo -e "To make this persistent, add this line to your ${BLUE}~/.bashrc${NC}, ${BLUE}~/.zshrc${NC}, or equivalent:"
echo -e "${BLUE}export PATH=\"$PWD/bin:\$PATH\"${NC}"
echo ""
echo -e "You can now use the following commands:"
echo -e "  ${GREEN}ai-tools${NC} - Main command with all tools"
echo -e "  ${GREEN}ai-tools typegen${NC} - Generate type definitions from API documentation"
echo -e "  ${GREEN}ai-tools docgen${NC} - Generate documentation for code"
echo -e "  ${GREEN}ai-tools-typegen${NC} - Standalone type generator"
echo -e "  ${GREEN}ai-tools-docgen${NC} - Standalone documentation generator"
echo ""
echo -e "Run ${BLUE}ai-tools --help${NC} to see all available options."
echo ""
echo -e "${RED}⚠ Remember to set your Gemini API key in the .env file!${NC}"
echo -e "${BLUE}=====================================${NC}"