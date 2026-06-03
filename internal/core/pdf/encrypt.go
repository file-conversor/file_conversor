// internal/pdf/encrypt.go

package pdf

import (
	"fmt"
	"os"
	"strings"

	"github.com/file-conversor/file_conversor/internal/core"
	"github.com/file-conversor/file_conversor/internal/core/flags"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/logger"
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

func (e EncryptionAlgorithm) String() string {
	switch e {
	case AES256:
		return "AES256"
	case AES128:
		return "AES128"
	case RC40:
		return "RC40"
	case RC128:
		return "RC128"
	default:
		return "Unknown"
	}
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

func (p *EncryptPermissions) String() string {
	perms := []string{}
	if p.PermissionAssemble {
		perms = append(perms, "Assemble")
	}
	if p.PermissionExtract {
		perms = append(perms, "Extract")
	}
	if p.PermissionPrint {
		perms = append(perms, "Print")
	}
	if p.PermissionModAnnFillForm {
		perms = append(perms, "Modify Annotations and Fill Forms")
	}
	if p.PermissionModify {
		perms = append(perms, "Modify")
	}
	if p.PermissionFill {
		perms = append(perms, "Fill Forms")
	}
	if p.PermissionAll {
		perms = append(perms, "All")
	}
	return fmt.Sprintf("[%s]", strings.Join(perms, ", "))
}

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
	return e.MapCommand.GetRunnable(func(inFile *os.File, outFile *os.File, runnable *core.Runnable, updateProgress interfaces.ProgressIncrement) error {
		// encrypt PDF from input file to output file
		conf := model.NewDefaultConfiguration()
		conf.OwnerPW = e.OwnerPassword
		conf.UserPW = e.UserPassword
		conf.Permissions = e.Permissions.Get()
		if err := e.Encryption.SetConf(conf); err != nil {
			return fmt.Errorf("set encryption configuration: %w", err)
		}
		logger.Infof(
			"Encrypting '%s' => '%s' (%s) with permissions: %s\n",
			inFile.Name(), outFile.Name(), e.Encryption.String(), e.Permissions.String(),
		)
		if err := api.Encrypt(inFile, outFile, conf); err != nil {
			return fmt.Errorf("pdf encrypt '%s' => '%s': %w", inFile.Name(), outFile.Name(), err)
		}
		return nil
	})
}
