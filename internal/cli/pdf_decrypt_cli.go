// internal/cli/pdf_decrypt_cli.go

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
// DECRYPT COMMAND
// -------------

type PdfDecryptCLI struct {
	InputFiles []string `arg:""    optional:""                    help:"Input PDF files."`
	OutputDir  string   `short:"o" required:""                    help:"Output directory for decrypted files."`
	Password   string   `short:"p" required:""                    help:"Password for decrypting PDF files."`
	Recurse    bool     `short:"r" optional:""                    help:"Recurse into subdirectories."`
	BatchFile  string   `short:"b" optional:""                    help:"Batch file with list of input files (one per line)."`
	Suffix     string   `short:"s" optional:"" default:"_decrypt" help:"Suffix to add to output file stem (default: '')"`
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
	// create decrypt function to run with or without progress bar
	cmd, err := pdf.NewDecrypt(
		c.Password,
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
	)
	if err != nil {
		return fmt.Errorf("pdf decrypt: %w", err)
	}

	// if no progress bars, just run the decrypt in a single thread and return any error
	logger.Infof("Decrypting input files into folder '%s'\n", c.OutputDir)
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
