// internal/cli/pdf_encrypt_cli.go

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
// ENCRYPT COMMAND
// -------------

type PdfEncryptionFlag struct {
	Encryption string `short:"e" optional:"" default:"aes256" enum:"aes256,aes128,rc128,rc40" help:"Encryption algorithm and key size."`
}

func (c *PdfEncryptionFlag) Get() pdf.EncryptionAlgorithm {
	switch c.Encryption {
	case "rc40":
		return pdf.RC40
	case "rc128":
		return pdf.RC128
	case "aes128":
		return pdf.AES128
	default:
		return pdf.AES256
	}
}

type PdfPermissionsFlag struct {
	Permissions []string `short:"m" optional:"" default:"all" enum:"all,print,modify,copy,annotate,none" help:"Permissions for encrypted PDF files (comma-separated)."`
}

func (c *PdfPermissionsFlag) Get() pdf.EncryptPermissions {
	perms := pdf.EncryptPermissions{}
	for _, perm := range c.Permissions {
		switch perm {
		case "print":
			perms.PermissionPrint = true
		case "modify":
			perms.PermissionModify = true
		case "copy":
			perms.PermissionExtract = true
		case "annotate":
			perms.PermissionModAnnFillForm = true
		case "none":
			// no permissions
		case "all":
			perms.PermissionAll = true
		}
	}
	return perms
}

type PdfEncryptCLI struct {
	InputFiles         []string `arg:""    optional:""                      help:"Input PDF files."`
	OutputDir          string   `short:"o" optional:"" default:"."          help:"Output directory for encrypted files."`
	OwnerPassword      string   `short:"p" required:""                      help:"Password for encrypting PDF files."`
	UserPassword       string   `short:"u" optional:""                      help:"User password for encrypted PDF files (leave empty to use owner password)."`
	Recurse            bool     `short:"r" optional:""                      help:"Recurse into subdirectories."`
	BatchFile          string   `short:"b" optional:""                      help:"Batch file with list of input files (one per line)."`
	Suffix             string   `short:"s" optional:"" default:"_encrypted" help:"Suffix to add to output file stem"`
	PdfEncryptionFlag           // encryption algorithm and key size for encrypting PDF files
	PdfPermissionsFlag          // permissions for encrypted PDF files
}

func (c *PdfEncryptCLI) Help() string {
	return `
Example usage:
  file_conversor pdf encrypt -p 1234 file1.pdf file2.pdf file3.pdf
  file_conversor pdf encrypt -p 1234 -o /path/to/output_directory -r /path/to/directory
  file_conversor pdf encrypt -p 1234 -u 5678 -o /path/to/output_directory -b /path/to/batchfile.txt
  cat file1.pdf | file_conversor pdf encrypt -p 1234 -o - > encrypted.pdf
`
}

func (c *PdfEncryptCLI) Run(ctx *MainCLI) error {
	// create encrypt function to run with or without progress bar
	encryptionAlgorithm := c.PdfEncryptionFlag.Get()
	permissions := c.PdfPermissionsFlag.Get()
	cmd, err := pdf.NewEncrypt(
		c.UserPassword,
		c.OwnerPassword,
		core.OutputDirFlag{
			OutputDir:    c.OutputDir,
			Overwrite:    ctx.Overwrite,
			Suffix:       c.Suffix,
			Format:       ".pdf",
			AcceptStdout: true,
		},
		core.InputsArg{
			InputFiles:  c.InputFiles,
			Recurse:     c.Recurse,
			BatchFile:   c.BatchFile,
			AcceptStdin: true,
		},
		encryptionAlgorithm,
		permissions,
	)
	if err != nil {
		return fmt.Errorf("pdf encrypt: %w", err)
	}

	// if no progress bars, just run the decrypt in a single thread and return any error
	logger.Infof("Encrypting input files into folder '%s'\n", c.OutputDir)
	if ctx.NoProgress {
		tp := env.NewThreadPool(0)
		for runnable := range cmd.GetRunnable() {
			tp.AddTask(runnable.Run)
		}
		return tp.Wait()
	}

	// progress bar with spinner style (since we don't know total pages in advance)
	p := progress.NewProgressBarMgr(0)
	for runnable := range cmd.GetRunnable() {
		barCfg := progress.NewBarCfg(env.BaseName(runnable.OutputPath), 0, true)
		p.AddBarOrSpinner(barCfg, runnable.Run)
	}
	return p.Wait()
}
