// internal/cli/pdf_merge_cli.go

package cli

import (
	"fmt"

	"github.com/file-conversor/file_conversor/internal/cli/progress"
	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/core/pdf"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/logger"
)

// -------------
// MERGE COMMAND
// -------------

type PdfMergeCLI struct {
	InputFiles []string `arg:""    optional:"" help:"Input PDF files (leave empty for stdin)."`
	OutputFile string   `short:"o" required:"" help:"Output PDF file (use - for stdout)."`
	Append     bool     `short:"a" optional:"" help:"Append to output file (no effect for stdout)."`
	Recurse    bool     `short:"r" optional:"" help:"Recurse into subdirectories."`
	BatchFile  string   `short:"b" optional:"" help:"Batch file with list of input files (one per line)."`
}

func (c *PdfMergeCLI) Help() string {
	return `
Example usage:
  file_conversor pdf merge -o merged.pdf file1.pdf file2.pdf file3.pdf
  file_conversor pdf merge -o merged.pdf -r /path/to/directory
  file_conversor pdf merge -o - file1.pdf file2.pdf file3.pdf > merged.pdf
  cat file1.pdf | file_conversor pdf merge -a -o merged.pdf
`
}

func (c *PdfMergeCLI) Run(ctx *MainCLI) error {
	// create merge function to run with or without progress bar
	mergeFunc := func() error {
		merge := &pdf.Merge{
			Append: c.Append,
			OutputFlag: core.OutputFlag{
				OutputFile: c.OutputFile,
				Overwrite:  ctx.Overwrite,
			},
			InputsArg: core.InputsArg{
				InputFiles: c.InputFiles,
				Recurse:    c.Recurse,
				BatchFile:  c.BatchFile,
			},
		}
		err := merge.Run()
		if err != nil {
			return fmt.Errorf("pdf merge - run: %w", err)
		}
		return nil
	}

	// if no progress bars, just run the merge in a single thread and return any error
	logger.Infof("Merging input files into '%s' (append: %t)\n", c.OutputFile, c.Append)
	if ctx.NoProgress {
		tp := env.NewThreadPool(0)
		tp.AddTask(mergeFunc)
		return tp.Wait()
	}

	// progress bar with spinner style (since we don't know total pages in advance)
	p := progress.NewProgressBarMgr(0)
	p.AddBarOrSpinner(progress.NewBarCfg(env.BaseName(c.OutputFile), 0, true),
		func(updateProgress progress.ProgressIncrement) error {
			return mergeFunc()
		})
	return p.Wait()
}
