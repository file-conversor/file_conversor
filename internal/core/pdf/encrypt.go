// internal/pdf/encrypt.go

package pdf

import (
	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/engine"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
)

type EncryptionAlgorithm = engine.PdfCpuEncryption
type EncryptPermissions = engine.PdfCpuPermissions

var (
	AES256 = engine.PdfCpuAES256
	AES128 = engine.PdfCpuAES128
	RC128  = engine.PdfCpuRC128
	RC40   = engine.PdfCpuRC40
)

type Encrypt struct {
	UserPassword    string              // user password for encryption (optional, if not set, will use owner password as user password)
	OwnerPassword   string              // owner password for encryption
	Encryption      EncryptionAlgorithm // encryption algorithm to use (e.g. AES256)
	Permissions     EncryptPermissions  // permissions for encrypted PDF
	core.MapCommand                     // MapCommand: batch input file => output directory
}

func NewEncrypt(
	userPassword, ownerPassword string,
	outputDirFlag flags.OutputDirFlag,
	inputsArg flags.InputFilesArg,
	encryption EncryptionAlgorithm,
	permissions EncryptPermissions,
) *Encrypt {
	return &Encrypt{
		UserPassword:  userPassword,
		OwnerPassword: ownerPassword,
		Encryption:    encryption,
		Permissions:   permissions,
		MapCommand: core.MapCommand{
			OutputDirFlag: outputDirFlag,
			InputFilesArg: inputsArg,
		},
	}
}

func (e *Encrypt) In() []string {
	return []string{".pdf"}
}

func (e *Encrypt) Out() []string {
	return []string{".pdf"}
}

func (e *Encrypt) Parse() error {
	if e.UserPassword == "" {
		e.UserPassword = e.OwnerPassword // if user password is not set, use owner password as user password
	}
	return e.MapCommand.Parse(e) // pass Encrypt as FormatInterface to MapCommand
}

func (e *Encrypt) Validate() error {
	return e.MapCommand.Validate(e) // pass Encrypt as FormatInterface to MapCommand
}

// Will run the encryption command and return any error encountered during execution.
func (e *Encrypt) GetRunnable() <-chan *core.Runnable {
	return e.MapCommand.GetRunnable(func(inFile string, outFile string, runnable *core.Runnable, updateProgress interfaces.ProgressIncrement) error {
		logger.Infof(
			"Encrypting '%s' => '%s' (%s) with permissions: %s\n",
			inFile, outFile, e.Encryption.String(), e.Permissions.String(),
		)
		pdfcpuEngine := engine.NewPdfCpuEngine(
			engine.NewPdfCpuPassword(e.OwnerPassword, e.UserPassword),
			e.Encryption,
			&e.Permissions,
			false,
		)
		return pdfcpuEngine.Encrypt(inFile, outFile)
	})
}
