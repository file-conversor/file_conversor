// internal/cli/pdf_encrypt_cli.go

package cli

import (
	cli_flags "github.com/file-conversor/file_conversor/internal/cli/flags"
	core_flags "github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/core/pdf"
	"github.com/file-conversor/file_conversor/internal/logger"
)

type PdfEncryptCLI struct {
	InputFiles                   []string `arg:""    optional:""                    help:"Input PDF files."`
	OutputDir                    string   `short:"o" optional:"" default:"."        help:"Output directory for encrypted files [default: ${default}]."`
	OwnerPassword                string   `short:"p" required:""                    help:"Password for encrypting PDF files."`
	UserPassword                 string   `short:"u" optional:""                    help:"User password for encrypted PDF files (leave empty to use owner password)."`
	Recurse                      bool     `short:"r" optional:""                    help:"Recurse into subdirectories."`
	BatchFile                    string   `short:"b" optional:""                    help:"Batch file with list of input files (one per line)."`
	Suffix                       string   `short:"s" optional:"" default:"_encrypt" help:"Suffix to add to output file stem [default: ${default}]."`
	cli_flags.PdfEncryptionFlag           // -e | encryption algorithm and key size for encrypting PDF files
	cli_flags.PdfPermissionsFlag          // -m | permissions for encrypted PDF files
}

func (c *PdfEncryptCLI) Help() string {
	return `
Example usage:
  file_conversor pdf encrypt -p 1234 file1.pdf file2.pdf file3.pdf
  file_conversor pdf encrypt -p 1234 -o /path/to/output_directory -r /path/to/directory
  file_conversor pdf encrypt -p 1234 -u 5678 -o /path/to/output_directory -b /path/to/batchfile.txt
`
}

func (c *PdfEncryptCLI) Run(ctx *MainCLI) error {
	cmd := pdf.NewEncrypt(
		c.UserPassword,
		c.OwnerPassword,
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
		c.PdfEncryptionFlag.Get(),
		c.PdfPermissionsFlag.Get(),
	)
	logger.Infof("Encrypting input files into folder '%s'\n", c.OutputDir)
	return ctx.ExecuteCmd(cmd, 0)
}
