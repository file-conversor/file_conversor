// internal/pdf/check.go

package pdf

import (
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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
	return d.FilterCommand.GetRunnable(func(inFile *os.File, runnable *core.Runnable, updateProgress interfaces.ProgressIncrement) error {
		logger.Infof("Checking '%s' ...\n", inFile.Name())
		conf := model.NewDefaultConfiguration()
		conf.UserPW = d.Password
		conf.OwnerPW = d.Password
		if err := api.Validate(inFile, conf); err != nil {
			return fmt.Errorf("check file '%s': %w", inFile.Name(), err)
		}
		return nil
	})
}
