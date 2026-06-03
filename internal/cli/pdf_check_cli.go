// internal/cli/pdf_check_cli.go

package cli

import (
	core_flags "github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/core/pdf"
	"github.com/file-conversor/file_conversor/internal/logger"
)

type PdfCheckCLI struct {
	InputFiles []string `arg:""    optional:""                    help:"Input PDF files."`
	Password   string   `short:"p" optional:""                    help:"Password for encrypted PDF files."`
	Recurse    bool     `short:"r" optional:""                    help:"Recurse into subdirectories."`
	BatchFile  string   `short:"b" optional:""                    help:"Batch file with list of input files (one per line)."`
}

func (c *PdfCheckCLI) Help() string {
	return `
Example usage:
  file_conversor pdf check file1.pdf file2.pdf file3.pdf
  file_conversor pdf check -r /path/to/directory
  file_conversor pdf check -b /path/to/batchfile.txt
  cat file1.pdf | file_conversor pdf check 
`
}

func (c *PdfCheckCLI) Run(ctx *MainCLI) error {
	// if no progress bars, just run the check in a single thread and return any error
	logger.Infof("Checking input files\n")
	cmd := pdf.NewCheck(
		c.Password,
		core_flags.InputFilesArg{
			InputFiles: c.InputFiles,
			Recurse:    c.Recurse,
			BatchFile:  c.BatchFile,
		},
	)
	return ctx.ExecuteCmd(cmd, 0)
}
