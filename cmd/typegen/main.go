package main

import (
	"log"
	"os"

	"github.com/kamdyn/ai-toolkit/pkg/common"
	"github.com/kamdyn/ai-toolkit/pkg/typegen"
	"github.com/urfave/cli/v2"
)

func main() {
	// Load environment variables
	common.LoadEnv()

	// Get the types generator command
	typeGenCmd := typegen.GetTypeGenCommand()

	// Create CLI app
	app := &cli.App{
		Name:    "ai-tools-typegen",
		Usage:   "Generate type definitions from API documentation",
		Version: common.Version,
		Flags:   typeGenCmd.Flags,
		Action:  typeGenCmd.Action,
		Before:  typeGenCmd.Before,
	}

	// Run the app
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}