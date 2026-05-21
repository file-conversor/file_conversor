// internal/cli/main.go

package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/logger"
)

const UndefErrorExitCode = 1
const PanicExitCode = 126
const LoggingCleanExitCode = 127

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
	var errGrp error = nil
	var cli MainCLI

	// parse CLI arguments
	ctx := kong.Parse(&cli,
		kong.Name(appName),
		kong.Description("Multi-format cross-platform file conversion and manipulation tool."),
		kong.ConfigureHelp(kong.HelpOptions{
			NoAppSummary:        false,              // show app usage summary
			Summary:             false,              // show all description in command
			Compact:             true,               // show flags for current command only
			Tree:                false,              // show command flat-list, instead of tree structure
			FlagsLast:           true,               // print flags after args
			Indenter:            kong.SpaceIndenter, // use spaces for indenting
			NoExpandSubcommands: true,               // don't expand subcommands in help
			WrapUpperBound:      0,                  // wrapping of help text
		}),
		kong.Exit(func(code int) {
			exitCode = code
			errGrp = fmt.Errorf("CLI parsing")
		}),
	)

	// setup logging
	stopLogging, errLog := initLogging(appName, cli.Verbose, cli.Quiet)
	if errLog != nil {
		errGrp = errors.Join(errGrp, fmt.Errorf("logging setup: %w", errLog))
	} else {
		defer func() {
			if err := stopLogging(); err != nil {
				err = fmt.Errorf("logging cleanup: %w", err)
				errGrp = errors.Join(errGrp, err)
				exitCode = LoggingCleanExitCode // special exit code for logging cleanup failure
				fmt.Fprintf(os.Stderr, "[ERROR] - %v", err)
			}
		}()
	}

	// run CLI command, if parsing was successful
	if errGrp == nil {
		errGrp = errors.Join(errGrp, ctx.Run(&cli))
	}

	// log any errors and ensure non-zero exit code if there was an error
	if errGrp != nil {
		logger.Errorf("%v\n", errGrp)
		if exitCode == 0 {
			logger.Errorf("Wrong exit code %d", exitCode)
			exitCode = UndefErrorExitCode // default error exit code
		}
	}
	return exitCode, errGrp
}

func initLogging(appName string, verbose bool, quiet bool) (logger.StopLoggerFunc, error) {
	noop := logger.DefaultStopLoggerFunc()
	// setup logging
	logfile, err := env.Logfile(appName)
	if err != nil {
		return noop, fmt.Errorf("logfile get: %w", err)
	}
	logCfg := logger.DefaultConfig(logfile)
	logMode := "Normal"
	if verbose {
		// set verbose logging to terminal
		logMode = "Verbose"
		logCfg.TerminalLevel = logger.DebugLevel
		fmt.Fprintf(os.Stderr, "[DEBUG] - Logfile: %s\n", logfile)
	}
	if quiet {
		// set quiet logging to terminal
		logMode = "Quiet"
		logCfg.TerminalLevel = logger.ErrorLevel
	}
	stopLogger, err := logger.SetupLogger(logCfg)
	if err != nil {
		return noop, fmt.Errorf("logger setup: %w", err)
	}
	logger.Debugf("Log mode: %s\n", logMode)
	return stopLogger, nil
}
