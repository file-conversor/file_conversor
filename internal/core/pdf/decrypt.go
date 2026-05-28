// internal/pdf/decrypt.go

package pdf

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type Decrypt struct {
	Password           string // password for decryption
	core.OutputDirFlag        // output dir flag (dir, overwrite, suffix, format)
	core.InputsArg            // input file arguments (paths, recurse flag, batch file)
}

func (d *Decrypt) In() []string {
	return []string{".pdf"}
}

func (d *Decrypt) Out() []string {
	return []string{".pdf"}
}

func NewDecrypt(password string, outputDirFlag core.OutputDirFlag, inputsArg core.InputsArg) (*Decrypt, error) {
	command := &Decrypt{
		Password:      password,
		OutputDirFlag: outputDirFlag,
		InputsArg:     inputsArg,
	}
	if err := command.Parse(); err != nil {
		return nil, err
	}
	if err := command.Validate(); err != nil {
		return nil, err
	}
	return command, nil
}

func (d *Decrypt) Parse() error {
	if err := errors.Join(
		d.OutputDirFlag.Parse(d),
		d.InputsArg.Parse(d),
	); err != nil {
		return fmt.Errorf("parse input files: %w", err)
	}
	return nil
}

func (d *Decrypt) Validate() error {
	return errors.Join(
		// validation for output file
		d.OutputDirFlag.Validate(d),

		// validation for input files
		d.InputsArg.Validate(d),
	)
}

// Will run the decryption command and return any error encountered during execution.
func (d *Decrypt) GetRunnable() <-chan core.Runnable {
	outChan := make(chan core.Runnable) // channel to send runnable to be executed

	go func() {
		defer close(outChan) // close channel when done

		processFileFunc := func(inputFile *os.File, cleanup func() error) error {
			runnable := core.NewRunnable()
			runnable.AppendCleanup(cleanup) // ensure cleanup is called after processing

			runnableRun := func(updateProgress interfaces.ProgressIncrement) error {
				// open output file
				outFile, err := d.GetOutputFile(inputFile.Name())
				if err != nil {
					return fmt.Errorf("open output file for '%s': %w", inputFile.Name(), err)
				}
				defer outFile.Close() // ensure output file is closed after processing
				runnable.SetOutputPath(outFile.Name())

				// check if PDF is encrypted before attempting decryption, if not,
				//     just copy the file to the output path
				isEncrypted, err := IsPdfEncrypted(inputFile.Name())
				if err != nil {
					return fmt.Errorf("check encryption status for '%s': %w", inputFile.Name(), err)
				}
				if !isEncrypted {
					logger.Warnf("pdf '%s' is not encrypted. Copying without decryption to '%s'...", inputFile.Name(), outFile.Name())
					_, err = io.Copy(outFile, inputFile)
					if err != nil {
						return fmt.Errorf("copy file '%s' to '%s': %w", inputFile.Name(), outFile.Name(), err)
					}
					return nil
				}

				// decrypt PDF from input file to output file
				conf := model.NewDefaultConfiguration()
				conf.OwnerPW = d.Password // pdfcpu chooses which password to use automatically
				conf.UserPW = d.Password
				if err := api.Decrypt(inputFile, outFile, conf); err != nil {
					runnable.AppendCleanup(func() error { return os.Remove(outFile.Name()) })
					return fmt.Errorf("pdf decrypt '%s' => '%s': %w", inputFile.Name(), outFile.Name(), err)
				}
				return nil
			}

			runnable.SetRun(runnableRun)
			outChan <- *runnable
			return nil
		}

		// process input files
		if err := env.OpenInputFiles(processFileFunc, d.InputFiles...); err != nil {
			runnable := core.NewRunnable()
			runnable.SetError(fmt.Errorf("pdf decrypt: %w", err))
			outChan <- *runnable
		}
	}()

	return outChan
}

func IsPdfEncrypted(inFile string) (bool, error) {
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
