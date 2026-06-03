// internal/cli/main.go

package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/file-conversor/file_conversor/internal/cli/progress"
	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/logger"
)

var terminalIo = env.NewIoProxy(os.Stderr)

const UndefErrorExitCode = 1
const CliParserExitCode = 80
const PanicExitCode = 126
const LoggingCleanExitCode = 127

type MainCLI struct {
	// Args:
	Pdf PdfCLI `cmd:"" help:"PDF conversion and manipulation commands."`

	// Flags :
	Quiet      bool `short:"Q" xor:"log" help:"Quiet output (show errors only)."`
	Debug      bool `short:"D" xor:"log" help:"Debug output (show debug information)."`
	NoProgress bool `short:"N"           help:"No progress bars or spinners."`
	Overwrite  bool `short:"O"           help:"Overwrite output files."`
	Install    bool `short:"I"           help:"Install dependencies, if needed (no user prompts)."`
}

func (c *MainCLI) Help() string {
	return `
Example usage:
  # process multiple inputs and write to output.pdf (no progress, overwrite output)
  file_conversor -N -O pdf merge -o output.pdf input1.pdf input2.pdf input3.pdf
`
}

func (c *MainCLI) ExecuteCmd(cmd core.CommandInterface, total int64) error {
	const maxWorkers = 0   // 0 means auto-detect and use all available CPU cores
	const progressFps = 15 // Hz (refresh rate) - 60 is too fast

	// parse and validate command before running any tasks to fail fast on invalid input
	if err := errors.Join(
		cmd.Parse(),
		cmd.Validate(),
	); err != nil {
		return err
	}

	// if no progress bars, just run tasks concurrently with a thread pool
	if c.NoProgress {
		tp := env.NewThreadPool(maxWorkers)
		for runnable := range cmd.GetRunnable() {
			tp.AddTask(runnable.Run)
		}
		return tp.Wait()
	}

	// progress bar
	pm := progress.NewProgressBarMgr(maxWorkers, progressFps)
	terminalIo.RouteTo(pm)              // route log output to progress manager to avoid control character issues
	defer terminalIo.RouteTo(os.Stderr) // ensure we stop routing log output when done
	for runnable := range cmd.GetRunnable() {
		barCfg := progress.NewBarCfg(env.BaseName(runnable.OutputPath), total, true)
		pm.AddBarOrSpinner(barCfg, runnable.Run)
	}
	return pm.Wait()
}

// Run executes the main CLI logic.
func Run(appName string) (int, error) {
	// default exit code is 0 (success)
	var terminate bool = false
	var exitCode int = 0
	var errGrp error = nil
	var cli MainCLI

	// force flush stdout, stderr
	defer os.Stdout.Sync()
	defer os.Stderr.Sync()

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
			terminate = true
			if code != 0 {
				exitCode = CliParserExitCode
				errGrp = fmt.Errorf("CLI parsing")
			}
		}),
	)
	// print a newline to separate error message
	if terminate {
		fmt.Fprintf(terminalIo, "\n")
	}

	// if if TTY is not available, disable progress bars and spinners
	// to avoid cluttering output with control characters
	if !env.IsTTY(os.Stdout) {
		cli.NoProgress = true
		cli.Quiet, cli.Debug = true, false
	}

	// setup logging
	logFile, errLog := initLogging(appName, cli.Debug, cli.Quiet)
	if errLog != nil {
		errGrp = errors.Join(errGrp, fmt.Errorf("logging setup: %w", errLog))
	} else {
		defer func() {
			if err := logFile.Close(); err != nil {
				err = fmt.Errorf("logging cleanup: %w", err)
				errGrp = errors.Join(errGrp, err)
				exitCode = LoggingCleanExitCode // special exit code for logging cleanup failure
				fmt.Fprintf(terminalIo, "[ERROR] - %v", err)
			}
		}()
	}

	// run CLI command, if parsing was successful
	if errGrp == nil && !terminate {
		errGrp = errors.Join(errGrp, ctx.Run(&cli))
	}

	// log any errors and ensure non-zero exit code if there was an error
	if errGrp != nil {
		logger.Errorf("%v\n", errGrp)
		if exitCode == 0 {
			exitCode = UndefErrorExitCode // default error exit code
		}
		logger.Errorf("Exit Code %d\n", exitCode)
	} else {
		logger.Debugf("Exit Code %d\n", exitCode)
	}
	return exitCode, errGrp
}

func initLogging(appName string, verbose bool, quiet bool) (io.Closer, error) {
	noop := io.NopCloser(nil)
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
		fmt.Fprintf(terminalIo, "[DEBUG] - Logfile: %s\n", logfile)
	}
	if quiet {
		// set quiet logging to terminal
		logMode = "Quiet"
		logCfg.TerminalLevel = logger.ErrorLevel
	}
	// set progress manager as terminal output to avoid control character issues
	logCfg.TerminalOutIo = terminalIo
	logFile, err := logger.SetupLogger(logCfg)
	if err != nil {
		return noop, fmt.Errorf("logger setup: %w", err)
	}
	logger.Debugf("Log mode: %s\n", logMode)
	return logFile, nil
}
