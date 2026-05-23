// internal/cli/pdf_merge_cli.go

package cli

import (
	"github.com/file-conversor/file_conversor/internal/cli/progress"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/logger"
	"github.com/file-conversor/file_conversor/internal/pdf"
)

// -------------
// MERGE COMMAND
// -------------

type PdfMergeCLI struct {
	Inputs  []string `arg:""    required:"" default:"-" help:"Input PDF files (leave empty for stdin)."`
	Output  string   `short:"o" required:""             help:"Output PDF file (use - for stdout)."`
	Append  bool     `short:"a" optional:""             help:"Append to output file (no effect for stdout)."`
	Recurse bool     `short:"r" optional:""             help:"Recurse into subdirectories."`
}

func (c *PdfMergeCLI) Help() string {
	return `
Example usage:
  # Merge multiple PDF files into one
  file_conversor pdf merge -o merged.pdf file1.pdf file2.pdf file3.pdf

  # Recursively append all PDF files in a directory into merged.pdf 
  file_conversor pdf merge -a -o merged.pdf -r /path/to/directory

  # Merge multiple PDF files and pipe output to stdout
  file_conversor pdf merge -o - file1.pdf file2.pdf file3.pdf > merged.pdf
  
  # Append PDF files from stdin to merged.pdf
  cat file1.pdf | file_conversor pdf merge -a -o merged.pdf
`
}

func (c *PdfMergeCLI) Run(ctx *MainCLI) error {
	// create merge function to run with or without progress bar
	mergeFunc := func() error {
		merge := &pdf.Merge{
			Overwrite: ctx.Overwrite,
			Append:    c.Append,
			Recurse:   c.Recurse,
			Output:    c.Output,
			Inputs:    c.Inputs,
		}
		return merge.Run()
	}

	// if no progress bars, just run the merge in a single thread and return any error
	logger.Infof("Merging input files into '%s' (append: %t)\n", c.Output, c.Append)
	if ctx.NoProgress {
		tp := env.NewThreadPool(0)
		tp.AddTask(mergeFunc)
		return tp.Wait()
	}

	// progress bar with spinner style (since we don't know total pages in advance)
	p := progress.NewProgressBarMgr(0)
	p.AddBarOrSpinner(progress.NewBarCfg(env.BaseName(c.Output), 0, true),
		func(updateProgress progress.ProgressIncrement) error {
			return mergeFunc()
		})
	return p.Wait()
}
