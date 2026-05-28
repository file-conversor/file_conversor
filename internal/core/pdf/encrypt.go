// internal/pdf/encrypt.go

package pdf

import (
	"errors"
	"fmt"
	"os"

	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type EncryptionAlgorithm uint8

const (
	AES256 EncryptionAlgorithm = iota
	AES128
	RC128
	RC40 // legacy
)

func (e EncryptionAlgorithm) SetConf(conf *model.Configuration) error {
	switch e {
	case AES256:
		conf.EncryptUsingAES = true
		conf.EncryptKeyLength = 256

	case AES128:
		conf.EncryptUsingAES = true
		conf.EncryptKeyLength = 128

	case RC40:
		conf.EncryptUsingAES = false
		conf.EncryptKeyLength = 40

	case RC128:
		conf.EncryptUsingAES = false
		conf.EncryptKeyLength = 128

	default:
		return fmt.Errorf("invalid encryption algorithm")
	}

	return nil
}

type EncryptPermissions struct {
	PermissionAssemble       bool // Assemble document (security handlers >= rev.3)
	PermissionExtract        bool // Copy, extract text & graphics
	PermissionPrint          bool // Print (security handlers rev.2), draft print (security handlers >= rev.3)
	PermissionModAnnFillForm bool // Add or modify annotations, fill form fields,
	PermissionModify         bool // Modify contents
	PermissionFill           bool // Fill existing form fields
	PermissionAll            bool // All permissions
}

func (p *EncryptPermissions) Get() model.PermissionFlags {
	var perms model.PermissionFlags = model.PermissionsNone
	if p.PermissionAssemble {
		perms |= model.PermissionAssembleRev3
	}
	if p.PermissionExtract {
		perms |= model.PermissionExtract | model.PermissionExtractRev3
	}
	if p.PermissionPrint {
		perms |= model.PermissionPrintRev2 | model.PermissionPrintRev3
	}
	if p.PermissionModAnnFillForm {
		perms |= model.PermissionModAnnFillForm
	}
	if p.PermissionModify {
		perms |= model.PermissionModify
	}
	if p.PermissionFill {
		perms |= model.PermissionFillRev3
	}
	if p.PermissionAll {
		perms |= model.PermissionsAll
	}
	return perms
}

type Encrypt struct {
	UserPassword       string              // user password for encryption (optional, if not set, will use owner password as user password)
	OwnerPassword      string              // owner password for encryption
	Encryption         EncryptionAlgorithm // encryption algorithm to use (e.g. AES256)
	Permissions        EncryptPermissions  // permissions for encrypted PDF
	core.OutputDirFlag                     // output dir flag (dir, overwrite, suffix, format)
	core.InputsArg                         // input file arguments (paths, recurse flag, batch file)
}

func (e *Encrypt) In() []string {
	return []string{".pdf"}
}

func (e *Encrypt) Out() []string {
	return []string{".pdf"}
}

func NewEncrypt(
	userPassword, ownerPassword string,
	outputDirFlag core.OutputDirFlag,
	inputsArg core.InputsArg,
	encryption EncryptionAlgorithm,
	permissions EncryptPermissions,
) (*Encrypt, error) {

	command := &Encrypt{
		UserPassword:  userPassword,
		OwnerPassword: ownerPassword,
		Encryption:    encryption,
		Permissions:   permissions,
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

func (e *Encrypt) Parse() error {
	if e.UserPassword == "" {
		e.UserPassword = e.OwnerPassword // if user password is not set, use owner password as user password
	}

	if err := errors.Join(
		e.OutputDirFlag.Parse(e),
		e.InputsArg.Parse(e),
	); err != nil {
		return fmt.Errorf("parse input files: %w", err)
	}
	return nil
}

func (e *Encrypt) Validate() error {
	return errors.Join(
		// validation for output file
		e.OutputDirFlag.Validate(e),

		// validation for input files
		e.InputsArg.Validate(e),
	)
}

// Will run the encryption command and return any error encountered during execution.
func (e *Encrypt) GetRunnable() <-chan core.Runnable {
	outChan := make(chan core.Runnable) // channel to send runnable to be executed

	go func() {
		defer close(outChan) // close channel when done

		processFileFunc := func(inputFile *os.File, cleanup func() error) error {
			runnable := core.NewRunnable()
			runnable.AppendCleanup(cleanup) // ensure cleanup is called after processing

			runnableRun := func(updateProgress interfaces.ProgressIncrement) error {
				// open output file
				outFile, err := e.GetOutputFile(inputFile.Name())
				if err != nil {
					return fmt.Errorf("open output file for '%s': %w", inputFile.Name(), err)
				}
				defer outFile.Close() // ensure output file is closed after processing
				runnable.SetOutputPath(outFile.Name())

				// encrypt PDF from input file to output file
				conf := model.NewDefaultConfiguration()
				conf.OwnerPW = e.OwnerPassword
				conf.UserPW = e.UserPassword
				conf.Permissions = e.Permissions.Get()
				if err := e.Encryption.SetConf(conf); err != nil {
					return fmt.Errorf("set encryption configuration: %w", err)
				}
				if err := api.Encrypt(inputFile, outFile, conf); err != nil {
					runnable.AppendCleanup(func() error { return os.Remove(outFile.Name()) })
					return fmt.Errorf("pdf encrypt '%s' => '%s': %w", inputFile.Name(), outFile.Name(), err)
				}
				return nil
			}

			runnable.SetRun(runnableRun)
			outChan <- *runnable
			return nil
		}

		// process input files
		if err := env.OpenInputFiles(processFileFunc, e.InputFiles...); err != nil {
			runnable := core.NewRunnable()
			runnable.SetError(fmt.Errorf("pdf encrypt: %w", err))
			outChan <- *runnable
		}
	}()

	return outChan
}
