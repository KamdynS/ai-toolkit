# AI Toolkit

A collection of AI-powered command-line tools for developers.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

AI Toolkit provides a set of command-line utilities powered by Google's Gemini AI to streamline common development tasks. Each tool is designed to do one thing well, with a consistent and simple interface.

Currently included tools:

- **typegen**: Generate type definitions from API documentation
- **docgen**: Generate documentation for code

## Installation

### Prerequisites

- Go 1.21 or later
- A Gemini API key from [Google AI Studio](https://ai.google.dev/)

### Quick Install

```bash
# Clone the repository
git clone https://github.com/kamdyn/ai-toolkit.git
cd ai-toolkit

# Install the toolkit
./install.sh

# Add to your PATH (temporary)
export PATH="$PWD/bin:$PATH"
```

The installer will:
1. Check for dependencies
2. Create a `.env` file if it doesn't exist
3. Build all the tools and place them in the `bin/` directory
4. Provide instructions for adding the tools to your PATH

### Getting an API Key

1. Visit [Google AI Studio](https://ai.google.dev/)
2. Create or sign in to your account
3. Go to "API Keys" and create a new key
4. Add your API key to the `.env` file:

```
GEMINI_API_KEY=your_api_key_here
```

## Usage

### All-in-One Command

The `ai-tools` command provides access to all tools:

```bash
# Show available tools
ai-tools --help

# Run typegen
ai-tools typegen --url="https://docs.stripe.com/api/charges"

# Run docgen
ai-tools docgen --file="path/to/file.js"
```

### Standalone Commands

Each tool is also available as a standalone command:

```bash
# Type generation
ai-tools-typegen --url="https://api-docs-example.com"

# Documentation generation
ai-tools-docgen --file="path/to/your/code.py"
```

### Global Options

All tools support these common options:

- `--api-key, -k`: Gemini API key (can also be set with GEMINI_API_KEY environment variable)
- `--model`: Gemini model to use (default: "gemini-2.0-flash")
- `--temp`: Temperature for generation (0.0-1.0) (default: 0.2)
- `--timeout`: Timeout in seconds (default: 120)
- `--verbose`: Enable verbose logging
- `--output, -o`: Output file path

## Tool: TypeGen

TypeGen scrapes API documentation websites and generates type definitions in various programming languages.

### Basic Usage

```bash
# Generate TypeScript types for a public API
ai-tools typegen --url="https://docs.stripe.com/api/charges"

# Specify output language
ai-tools typegen --url="https://docs.stripe.com/api/charges" --lang=python

# Focus on a specific function/method
ai-tools typegen --url="https://docs.openai.com/api-reference/chat" --func="createChatCompletion"

# Specify an output file
ai-tools typegen --url="https://docs.github.com/en/rest/issues/issues" --output=github-issues.d.ts
```

### Supported Languages

- TypeScript/JavaScript
- Go
- Python
- Rust
- Java
- C#
- Swift
- Kotlin

## Tool: DocGen

DocGen generates comprehensive documentation for code files, supporting automatic language detection and various documentation styles.

### Basic Usage

```bash
# Generate documentation for a file (auto-detects language from extension)
ai-tools docgen --file=path/to/your/code.js

# Specify documentation style
ai-tools docgen --file=path/to/your/code.go --style=godoc

# Specify output file
ai-tools docgen --file=path/to/your/code.py --output=documented-code.py

# Generate with a specific model
ai-tools docgen --file=path/to/your/code.js --model="gemini-2.0-pro"
```

### Documentation Styles

- **jsdoc**: JavaScript/TypeScript JSDoc style
- **godoc**: Go documentation style
- **docstring**: Python docstring style (Google format)
- **xml**: XML documentation style (C#/Java)
- **markdown**: Markdown documentation

### Project Documentation

The `--dir` flag enables comprehensive documentation for all source files in a project:

```bash
# Generate documentation for the current directory
ai-tools docgen --dir=.

# Generate documentation with custom title
ai-tools docgen --dir=./my-project --title="My Amazing Project"

# Specify output location
ai-tools docgen --dir=./my-project --output=./docs/PROJECT.md
```

This will:
1. Generate individual documentation files for each source file in a `docs/` directory
2. Create a combined `PROJECT.md` file with:
   - Table of contents
   - Project overview
   - Documentation for all files, organized by directory

## Environment Variables

You can set default values in the `.env` file:

```
# Required
GEMINI_API_KEY=your_gemini_api_key_here

# Optional defaults
DEFAULT_LANG=typescript
DEFAULT_MODEL=gemini-2.0-flash
DEFAULT_TEMPERATURE=0.2
DEFAULT_TIMEOUT=120
DEFAULT_VERBOSE=false
```

Examples can additionally be found in `.env.example`

## Contributing

Contributions are welcome! Feel free to:

1. Open issues for feature requests or bugs
   1. This is actually the preferred way to recommend a new tool to add to this CLI. 
2. Submit pull requests for new tools or improvements
   1. I will accept a PR of a new feature without opening a discussion if it is fully formed. :)
3. Suggest new AI-powered developer tools that would be useful

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.