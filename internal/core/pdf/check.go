// internal/pdf/check.go

package pdf

import (
	"fmt"

	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/engine"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
)

type Check struct {
	Password           string `short:"p" help:"Password for encrypted PDF files."`
	core.FilterCommand        // MapCommand: batch input file => output directory
}

func NewCheck(password string, inputsArg flags.InputFilesArg) *Check {
	return &Check{
		Password: password,
		FilterCommand: core.FilterCommand{
			InputFilesArg: inputsArg,
		},
	}
}

func (d *Check) In() []string {
	return []string{".pdf"}
}

func (d *Check) Out() []string {
	return []string{".pdf"}
}

func (d *Check) Parse() error {
	return d.FilterCommand.Parse(d) // pass Check as FormatInterface to FilterCommand
}

func (d *Check) Validate() error {
	return d.FilterCommand.Validate(d) // pass Check as FormatInterface to FilterCommand
}

func (d *Check) GetRunnable() <-chan *core.Runnable {
	return d.FilterCommand.GetRunnable(func(inFile string, runnable *core.Runnable, updateProgress interfaces.ProgressIncrement) error {
		logger.Infof("Checking '%s' ...\n", inFile)
		pdfcpuEngine := engine.NewPdfCpuEngine(
			engine.NewPdfCpuPassword(d.Password, d.Password),
			engine.PdfCpuEncryptionNone,
			engine.PdfCpuPermissionsNone,
			false,
		)
		if err := pdfcpuEngine.Check(inFile); err != nil {
			return fmt.Errorf("check file '%s': %w", inFile, err)
		}
		return nil
	})
}
