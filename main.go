package main

import (
	"log"
	"os"

	"github.com/kamdyn/ai-toolkit/pkg/common"
	"github.com/kamdyn/ai-toolkit/pkg/docgen"
	"github.com/kamdyn/ai-toolkit/pkg/typegen"
	"github.com/urfave/cli/v2"
)

func main() {
	// Load environment variables
	common.LoadEnv()

	// Create CLI app
	app := &cli.App{
		Name:    "ai-tools",
		Usage:   "A collection of AI-powered tools for developers",
		Version: common.Version,
		Commands: []*cli.Command{
			{
				Name:    "typegen",
				Aliases: []string{"types", "t"},
				Usage:   "Generate type definitions from API documentation",
				Flags:   typegen.GetTypeGenCommand().Flags,
				Action:  typegen.GetTypeGenCommand().Action,
				Before:  typegen.GetTypeGenCommand().Before,
			},
			{
				Name:    "docgen",
				Aliases: []string{"docs", "d"},
				Usage:   "Generate documentation for code",
				Flags:   docgen.GetDocGenCommand().Flags,
				Action:  docgen.GetDocGenCommand().Action,
				Before:  docgen.GetDocGenCommand().Before,
			},
		},
	}

	// Run the app
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}