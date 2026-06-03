// internal/pdf/decrypt.go

package pdf

import (
	"fmt"

	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/engine"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
)

type Decrypt struct {
	Password        string // password for decryption
	core.MapCommand        // MapCommand: batch input file => output directory
}

func NewDecrypt(password string, outputDirFlag flags.OutputDirFlag, inputsArg flags.InputFilesArg) *Decrypt {
	return &Decrypt{
		Password: password,
		MapCommand: core.MapCommand{
			OutputDirFlag: outputDirFlag,
			InputFilesArg: inputsArg,
		},
	}
}

func (d *Decrypt) In() []string {
	return []string{".pdf"}
}

func (d *Decrypt) Out() []string {
	return []string{".pdf"}
}

func (d *Decrypt) Parse() error {
	return d.MapCommand.Parse(d) // pass Decrypt as FormatInterface to MapCommand
}

func (d *Decrypt) Validate() error {
	return d.MapCommand.Validate(d) // pass Decrypt as FormatInterface to MapCommand
}

func (d *Decrypt) GetRunnable() <-chan *core.Runnable {
	return d.MapCommand.GetRunnable(func(inFile string, outFile string, runnable *core.Runnable, updateProgress interfaces.ProgressIncrement) error {
		pdfcpuEngine := engine.NewPdfCpuEngine(
			engine.NewPdfCpuPassword(d.Password, d.Password),
			engine.PdfCpuEncryptionNone,
			engine.PdfCpuPermissionsNone,
			false,
		)
		isEncrypted, err := pdfcpuEngine.IsPdfEncrypted(inFile)
		if err != nil {
			return fmt.Errorf("check encryption status for '%s': %w", inFile, err)
		}
		if !isEncrypted {
			logger.Warnf("PDF '%s' not encrypted. Copying as-is to '%s'...\n", inFile, outFile)
			err := env.CopyFile(inFile, outFile)
			if err != nil {
				return fmt.Errorf("copy file '%s' to '%s': %w", inFile, outFile, err)
			}
			return nil
		}

		logger.Infof("Decrypting '%s' => '%s' ...\n", inFile, outFile)
		return pdfcpuEngine.Decrypt(inFile, outFile)
	})
}
