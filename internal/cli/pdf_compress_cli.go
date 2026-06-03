// internal/cli/pdf_compress_cli.go

package cli

import (
	core_flags "github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/core/pdf"
	"github.com/file-conversor/file_conversor/internal/logger"
)

type PdfCompressCLI struct {
	InputFiles []string `arg:""    optional:""                    help:"Input PDF files."`
	OutputDir  string   `short:"o" optional:"" default:"."        help:"Output directory [default: ${default}]."`
	Recurse    bool     `short:"r" optional:""                    help:"Recurse into subdirectories."`
	BatchFile  string   `short:"b" optional:""                    help:"Batch file with list of input files (one per line)."`
	Suffix     string   `short:"s" optional:"" default:"_compress" help:"Suffix to add to output file stem [default: ${default}]."`
}

func (c *PdfCompressCLI) Help() string {
	return `
Example usage:
  file_conversor pdf compress -o /path/to/output directory file1.pdf file2.pdf file3.pdf
  file_conversor pdf compress -o /path/to/output directory -r /path/to/directory
  file_conversor pdf compress -o /path/to/output directory -b /path/to/batchfile.txt
`
}

func (c *PdfCompressCLI) Run(ctx *MainCLI) error {
	cmd := pdf.NewCompress(
		core_flags.OutputDirFlag{
			OutputDir: c.OutputDir,
			Overwrite: ctx.Overwrite,
			Suffix:    c.Suffix,
			Format:    ".pdf",
		},
		core_flags.InputFilesArg{
			InputFiles: c.InputFiles,
			Recurse:    c.Recurse,
			BatchFile:  c.BatchFile,
		},
	)
	logger.Infof("Compressing input files into folder '%s'\n", c.OutputDir)
	return ctx.ExecuteCmd(cmd, 0)
}
