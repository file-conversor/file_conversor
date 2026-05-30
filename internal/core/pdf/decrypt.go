// internal/pdf/decrypt.go

package pdf

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type Decrypt struct {
	Password        string // password for decryption
	core.MapCommand        // MapCommand: batch input file => output directory
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

func NewDecrypt(password string, outputDirFlag flags.OutputDirFlag, inputsArg flags.InputFilesArg) (*Decrypt, error) {
	command := &Decrypt{
		Password: password,
		MapCommand: core.MapCommand{
			OutputDirFlag: outputDirFlag,
			InputFilesArg: inputsArg,
		},
	}
	if err := errors.Join(
		command.Parse(),
		command.Validate(),
	); err != nil {
		return nil, err
	}
	return command, nil
}

func (d *Decrypt) GetRunnable() <-chan *core.Runnable {
	return d.MapCommand.GetRunnable(func(inFile *os.File, outFile *os.File, runnable *core.Runnable, updateProgress interfaces.ProgressIncrement) error {
		// check if PDF is encrypted before attempting decryption, if not,
		//     just copy the file to the output path
		isEncrypted, err := d.IsPdfEncrypted(inFile.Name())
		if err != nil {
			return fmt.Errorf("check encryption status for '%s': %w", inFile.Name(), err)
		}
		if !isEncrypted {
			logger.Warnf("PDF '%s' is not encrypted. Copying as-is to '%s'...", inFile.Name(), outFile.Name())
			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return fmt.Errorf("copy file '%s' to '%s': %w", inFile.Name(), outFile.Name(), err)
			}
			return nil
		}

		// decrypt PDF from input file to output file
		conf := model.NewDefaultConfiguration()
		conf.OwnerPW = d.Password // pdfcpu chooses which password to use automatically
		conf.UserPW = d.Password
		logger.Infof(
			"Decrypting '%s' => '%s'\n",
			inFile.Name(), outFile.Name(),
		)
		if err := api.Decrypt(inFile, outFile, conf); err != nil {
			return fmt.Errorf("pdf decrypt '%s' => '%s': %w", inFile.Name(), outFile.Name(), err)
		}
		return nil
	})
}

func (d *Decrypt) IsPdfEncrypted(inFile string) (bool, error) {
	ctx, err := api.ReadContextFile(inFile)
	if errors.Is(err, pdfcpu.ErrWrongPassword) {
		return true, nil // if wrong password error, then PDF is encrypted
	}
	if err != nil {
		return false, err
	}

	if ctx.Encrypt != nil {
		return true, nil
	}
	return false, nil
}
