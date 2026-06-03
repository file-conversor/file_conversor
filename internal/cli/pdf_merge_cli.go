// internal/cli/pdf_merge_cli.go

package cli

import (
	core_flags "github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/core/pdf"
	"github.com/file-conversor/file_conversor/internal/logger"
)

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
`
}

func (c *PdfMergeCLI) Run(ctx *MainCLI) error {
	cmd := pdf.NewMerge(
		c.Append,
		core_flags.OutputFileFlag{
			OutputFile: c.OutputFile,
			Overwrite:  ctx.Overwrite,
		},
		core_flags.InputFilesArg{
			InputFiles: c.InputFiles,
			Recurse:    c.Recurse,
			BatchFile:  c.BatchFile,
		},
	)
	logger.Infof("Merging input files into '%s' (append: %t)\n", c.OutputFile, c.Append)
	return ctx.ExecuteCmd(cmd, 0)
}
