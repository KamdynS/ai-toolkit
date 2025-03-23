package main

import (
	"log"
	"os"

	"github.com/kamdyn/ai-toolkit/pkg/common"
	"github.com/kamdyn/ai-toolkit/pkg/docgen"
	"github.com/urfave/cli/v2"
)

func main() {
	// Load environment variables
	common.LoadEnv()

	// Get the docgen command
	docgenCmd := docgen.GetDocGenCommand()

	// Create CLI app
	app := &cli.App{
		Name:    "ai-tools-docgen",
		Usage:   "Generate documentation for code",
		Version: common.Version,
		Flags:   docgenCmd.Flags,
		Action:  docgenCmd.Action,
		Before:  docgenCmd.Before,
	}

	// Run the app
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}