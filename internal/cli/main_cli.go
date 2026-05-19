// internal/cli/main.go

package cli

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/logger"
)

type MainCLI struct {
	// Args:
	Pdf PdfCLI `cmd:"" help:"PDF conversion and manipulation commands."`

	// Flags :
	Quiet     bool `short:"q" help:"Quiet output (show errors only)."`
	Verbose   bool `short:"v" help:"Verbose output (show debug information)."`
	Overwrite bool `short:"O" help:"Overwrite output files."`
	Install   bool `short:"I" help:"Install dependencies, if needed (no user prompts)."`
}

// Run executes the main CLI logic.
func Run(appName string) (int, error) {
	// default exit code is 0 (success)
	var exitCode int = 0

	// parse CLI arguments
	var cli MainCLI
	ctx := kong.Parse(&cli,
		kong.Name(appName),
		kong.Description("Multi-format cross-platform file conversion and manipulation tool."),
		kong.ConfigureHelp(kong.HelpOptions{
			NoAppSummary:        false,              // show app usage summary
			Summary:             true,               // show one-line summary of subcommands
			Compact:             true,               // show flags for current command only
			Tree:                false,              // show command flat-list, instead of tree structure
			FlagsLast:           true,               // print flags after args
			Indenter:            kong.SpaceIndenter, // use spaces for indenting
			NoExpandSubcommands: true,               // don't expand subcommands in help
			WrapUpperBound:      0,                  // wrapping of help text
		}),
		kong.UsageOnError(),
		kong.Exit(func(code int) {
			exitCode = code
		}),
	)
	if exitCode != 0 {
		return exitCode, fmt.Errorf("CLI parsing failed")
	}

	// setup logging
	stopLogging, err := initLogging(appName, &cli)
	if err != nil {
		return 1, fmt.Errorf("init logging: %w", err)
	}
	defer stopLogging()

	// run CLI
	return exitCode, ctx.Run(&cli)
}

func initLogging(appName string, ctx *MainCLI) (func(), error) {
	noop := func() {}
	// setup logging
	logfile, err := env.Logfile(appName)
	if err != nil {
		return noop, fmt.Errorf("logfile get: %w", err)
	}
	logCfg := logger.DefaultConfig(logfile)
	logMode := "Normal"
	if ctx.Verbose {
		// set verbose logging to terminal
		logMode = "Verbose"
		logCfg.TerminalLevel = logger.DebugLevel
	}
	if ctx.Quiet {
		// set quiet logging to terminal
		logMode = "Quiet"
		logCfg.TerminalLevel = logger.ErrorLevel
	}
	file, err := logger.SetupLogger(logCfg)
	if err != nil {
		return noop, fmt.Errorf("logger setup: %w", err)
	}
	logger.Debugf("Log mode: %s\n", logMode)
	return func() {
		file.Close()
	}, nil
}
