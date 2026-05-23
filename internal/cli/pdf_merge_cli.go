// internal/cli/pdf_merge_cli.go

package cli

import (
	"errors"

	"github.com/file-conversor/file_conversor/internal/cli/progress"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/logger"
	"github.com/file-conversor/file_conversor/internal/pdf"
	"github.com/file-conversor/file_conversor/internal/validation"
)

// -------------
// MERGE COMMAND
// -------------

type PdfMergeCLI struct {
	Inputs []string `arg:""    required:"" default:"-" help:"Input PDF files (leave empty for stdin)."`
	Output string   `short:"o" required:""             help:"Output PDF file (use - for stdout)."`
	Append bool     `short:"a" optional:""             help:"Append to output file (not supported for stdout)."`
}

func (c *PdfMergeCLI) Validate(ctx *MainCLI) error {
	if err := errors.Join(
		// validation for output file
		validation.CheckOutputStdout(true, c.Output),
		validation.OutputFileExt(c.Output, ".pdf"),
		validation.OutputFileOverwritable(c.Output, ctx.Overwrite || c.Append),

		// validation for input files
		validation.CheckInputStdin(true, c.Inputs...),
		validation.InputFileExt(c.Inputs, ".pdf"),
		validation.InputFileExists(c.Inputs...),

		// validation for input and output (together)
		validation.InputOutputNotEqual(c.Output, c.Inputs...),
	); err != nil {
		return err
	}
	return nil
}

func (c *PdfMergeCLI) Run(ctx *MainCLI) error {
	// parse and validate arguments and flags
	if err := c.Validate(ctx); err != nil {
		return err
	}

	// if no progress bars, just run the merge in a single thread and return any error
	logger.Infof("Merging input files into '%s' (append: %t)\n", c.Output, c.Append)
	if ctx.NoProgress {
		tp := env.NewThreadPool(0)
		tp.AddTask(func() error {
			return pdf.Merge(c.Append, c.Output, c.Inputs...)
		})
		return tp.Wait()
	}

	// progress bar with spinner style (since we don't know total pages in advance)
	p := progress.NewProgressBarMgr(0)
	p.AddBarOrSpinner(progress.NewBarCfg(env.BaseName(c.Output), 0, true),
		func(updateProgress progress.ProgressIncrement) error {
			return pdf.Merge(c.Append, c.Output, c.Inputs...)
		})
	return p.Wait()
}
