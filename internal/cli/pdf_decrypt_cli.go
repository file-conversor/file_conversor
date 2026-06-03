// internal/cli/pdf_decrypt_cli.go

package cli

import (
	core_flags "github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/core/pdf"
	"github.com/file-conversor/file_conversor/internal/logger"
)

type PdfDecryptCLI struct {
	InputFiles []string `arg:""    optional:""                    help:"Input PDF files (leave empty for stdin)."`
	OutputDir  string   `short:"o" optional:"" default:"."        help:"Output directory (use - for stdout) [default: ${default}]."`
	Password   string   `short:"p" required:""                    help:"Password for decrypting."`
	Recurse    bool     `short:"r" optional:""                    help:"Recurse into subdirectories."`
	BatchFile  string   `short:"b" optional:""                    help:"Batch file with list of input files (one per line)."`
	Suffix     string   `short:"s" optional:"" default:"_decrypt" help:"Suffix to add to output file stem [default: ${default}]."`
}

func (c *PdfDecryptCLI) Help() string {
	return `
Example usage:
  file_conversor pdf decrypt -o /path/to/output directory file1.pdf file2.pdf file3.pdf
  file_conversor pdf decrypt -o /path/to/output directory -r /path/to/directory
  file_conversor pdf decrypt -o /path/to/output directory -b /path/to/batchfile.txt
  cat file1.pdf | file_conversor pdf decrypt -o - > decrypted.pdf
`
}

func (c *PdfDecryptCLI) Run(ctx *MainCLI) error {
	cmd := pdf.NewDecrypt(
		c.Password,
		core_flags.OutputDirFlag{
			OutputDir:    c.OutputDir,
			Overwrite:    ctx.Overwrite,
			Suffix:       c.Suffix,
			Format:       ".pdf",
			AcceptStdout: true,
		},
		core_flags.InputFilesArg{
			InputFiles:  c.InputFiles,
			Recurse:     c.Recurse,
			BatchFile:   c.BatchFile,
			AcceptStdin: true,
		},
	)
	logger.Infof("Decrypting input files into folder '%s'\n", c.OutputDir)
	return ctx.ExecuteCmd(cmd, 0)
}
