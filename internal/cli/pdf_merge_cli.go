// internal/cli/pdf_merge_cli.go

package cli

import (
	"errors"
	"fmt"

	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/logger"
	"github.com/file-conversor/file_conversor/internal/pdf"
	"github.com/file-conversor/file_conversor/internal/progress"
	"github.com/file-conversor/file_conversor/internal/validation"
)

// -------------
// MERGE COMMAND
// -------------

type PdfMergeCLI struct {
	Inputs []string `arg:""    required:"" type:"existingfile" help:"Input PDF files."`
	Output string   `short:"o" required:"" type:"path"         help:"Output PDF file."`
	Append bool     `short:"a"                                 help:"Append to output file (no overwrite)."`
}

func (c *PdfMergeCLI) Validate(ctx *MainCLI) error {
	if err := errors.Join(
		validation.OutputFileOverwritable(c.Output, ctx.Overwrite || c.Append),
		validation.OutputFileExt(c.Output, ".pdf"),
		validation.InputFileExt(c.Inputs, ".pdf"),
		validation.InputOutputNotEqual(c.Output, c.Inputs...),
	); err != nil {
		return err
	}
	return nil
}

func (c *PdfMergeCLI) Run(ctx *MainCLI) error {
	// validate arguments and flags
	if err := c.Validate(ctx); err != nil {
		return err
	}

	// ensure output directory exists
	if err := env.EnsureParentDirExists(c.Output); err != nil {
		return fmt.Errorf("output dir cannot be created: %w", err)
	}

	// if no progress bars, just run the merge in a single thread and return any error
	logger.Infof("Merging input files into '%s' (append: %t)\n", c.Output, c.Append)
	if ctx.NoProgress {
		tp := env.NewThreadPool(0)
		tp.AddTask(func() error {
			return pdf.Merge(c.Output, c.Inputs, c.Append)
		})
		return tp.Wait()
	}

	// progress bar with spinner style (since we don't know total pages in advance)
	p := progress.NewProgressBarMgr()
	p.AddBarOrSpinner(progress.NewBarCfg(env.BaseName(c.Output), 0),
		func(updateProgress progress.ProgressIncrement) error {
			return pdf.Merge(c.Output, c.Inputs, c.Append)
		})
	return p.Wait()
}
