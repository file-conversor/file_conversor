// internal/cli/pdf_merge_cli.go

package cli

import (
	"fmt"

	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/pdf"
	"github.com/file-conversor/file_conversor/internal/utils"
)

// -------------
// MERGE COMMAND
// -------------

type PdfMergeCLI struct {
	Inputs []string `arg:""    required:"" type:"existingfile" help:"Input PDF files."`
	Output string   `short:"o" required:"" type:"path"         help:"Output PDF file."`
	Append bool     `short:"a"                                 help:"Append to output file (no overwrite)."`
}

func (c *PdfMergeCLI) AfterApply() error {
	if !env.IsFileExt(c.Output, ".pdf") {
		return fmt.Errorf("output file not a PDF: %s", c.Output)
	}
	for _, input := range c.Inputs {
		if !env.IsFileExt(input, ".pdf") {
			return fmt.Errorf("input file not a PDF: %s", input)
		}
		isequal, err := env.IsEqualPath(input, c.Output)
		if err != nil || isequal {
			return fmt.Errorf("input == output: %w", err)
		}
	}
	return nil
}

func (c *PdfMergeCLI) Run(ctx *MainCLI) error {
	if env.FileExists(c.Output) && !ctx.Overwrite && !c.Append {
		return fmt.Errorf("output file already exists: %s", c.Output)
	}
	if err := env.EnsureParentDirExists(c.Output); err != nil {
		return fmt.Errorf("output dir cannot be created: %w", err)
	}
	p := utils.NewProgressBarMgr()
	p.AddBarOrSpinner(utils.NewBarCfg(env.BaseName(c.Output), 0),
		func(updateProgress utils.ProgressIncrement) error {
			return pdf.Merge(c.Output, c.Inputs, c.Append)
		})
	return p.Wait()
}
