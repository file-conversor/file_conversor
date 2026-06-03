// internal/engine/pdfcpu.go

package engine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type PdfCpuEncryption int

const (
	PdfCpuEncryptionNone PdfCpuEncryption = iota
	PdfCpuRC40
	PdfCpuRC128
	PdfCpuAES128
	PdfCpuAES256
)

func (e *PdfCpuEncryption) SetConfig(conf *model.Configuration) {
	switch *e {
	case PdfCpuEncryptionNone:
		// no encryption settings needed
	case PdfCpuRC40:
		conf.EncryptUsingAES = false
		conf.EncryptKeyLength = 40
	case PdfCpuRC128:
		conf.EncryptUsingAES = false
		conf.EncryptKeyLength = 128
	case PdfCpuAES128:
		conf.EncryptUsingAES = true
		conf.EncryptKeyLength = 128
	default: // default to AES-256 if unrecognized or not specified
		conf.EncryptUsingAES = true
		conf.EncryptKeyLength = 256
	}
}

func (e *PdfCpuEncryption) String() string {
	switch *e {
	case PdfCpuEncryptionNone:
		return "No Encryption"
	case PdfCpuRC40:
		return "RC40"
	case PdfCpuRC128:
		return "RC128"
	case PdfCpuAES128:
		return "AES128"
	default:
		return "AES256"
	}
}

// ================
//  PASSWORD
// ================

type PdfCpuPassword struct {
	Owner string
	User  string
}

var (
	PdfCpuPasswordEmpty = &PdfCpuPassword{}
)

func NewPdfCpuPassword(ownerPw string, userPw string) *PdfCpuPassword {
	return &PdfCpuPassword{
		Owner: ownerPw,
		User:  userPw,
	}
}

func (p *PdfCpuPassword) SetConfig(conf *model.Configuration) {
	conf.OwnerPW = p.Owner
	conf.UserPW = p.User
}

// ===========
// PERMISSIONS
// ===========

type PdfCpuPermissions struct {
	PermissionAssemble       bool // Assemble document (security handlers >= rev.3)
	PermissionExtract        bool // Copy, extract text & graphics
	PermissionPrint          bool // Print (security handlers rev.2), draft print (security handlers >= rev.3)
	PermissionModAnnFillForm bool // Add or modify annotations, fill form fields,
	PermissionModify         bool // Modify contents
	PermissionFill           bool // Fill existing form fields
	PermissionAll            bool // All permissions
}

var (
	PdfCpuPermissionsNone = &PdfCpuPermissions{}
	PdfCpuPermissionsAll  = &PdfCpuPermissions{PermissionAll: true}
)

func (p *PdfCpuPermissions) SetConfig(config *model.Configuration) {
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
	// set config permissions
	config.Permissions = perms
}

func (p *PdfCpuPermissions) String() string {
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

// =========
//   ENGINE
// =========

type PdfCpuEngine struct {
	conf *model.Configuration
}

func NewPdfCpuEngine(
	password *PdfCpuPassword,
	encryption PdfCpuEncryption,
	permissions *PdfCpuPermissions,
	optimize bool,
) *PdfCpuEngine {
	// set configuration
	conf := model.NewDefaultConfiguration()

	// Encryption settings
	password.SetConfig(conf)
	encryption.SetConfig(conf)
	permissions.SetConfig(conf)

	// Optimization settings
	conf.Optimize = optimize                        // optimize pdfs
	conf.OptimizeBeforeWriting = optimize           // perform optimization before writing output (add optimize before final operation)
	conf.OptimizeDuplicateContentStreams = optimize // Deduplicates content streams (e.g. images) across pages
	conf.OptimizeResourceDicts = optimize           // Deduplicates and compresses resource dictionaries (e.g. fonts, images) across pages

	return &PdfCpuEngine{
		conf: conf,
	}
}

func (e *PdfCpuEngine) Check(inFile string) error {
	if err := api.ValidateFile(inFile, e.conf); err != nil {
		return fmt.Errorf("pdfcpu check file '%s': %w", inFile, err)
	}
	return nil
}

func (e *PdfCpuEngine) Compress(inFile string, outFile string) error {
	if err := api.OptimizeFile(inFile, outFile, e.conf); err != nil {
		return fmt.Errorf("pdfcpu compress '%s' => '%s': %w", inFile, outFile, err)
	}
	return nil
}

func (e *PdfCpuEngine) Decrypt(inFile string, outFile string) error {
	if err := api.DecryptFile(inFile, outFile, e.conf); err != nil {
		return fmt.Errorf("pdfcpu decrypt '%s' => '%s': %w", inFile, outFile, err)
	}
	return nil
}

func (e *PdfCpuEngine) Encrypt(inFile string, outFile string) error {
	if err := api.EncryptFile(inFile, outFile, e.conf); err != nil {
		return fmt.Errorf("pdfcpu encrypt '%s' => '%s': %w", inFile, outFile, err)
	}
	return nil
}

func (e *PdfCpuEngine) IsPdfEncrypted(inFile string) (bool, error) {
	ctx, err := api.ReadContextFile(inFile)
	if errors.Is(err, pdfcpu.ErrWrongPassword) {
		return true, nil // if wrong password error, then PDF is encrypted
	}
	if err != nil {
		return false, fmt.Errorf("pdfcpu is encrypted for '%s': %w", inFile, err)
	}
	return (ctx.Encrypt != nil), nil
}

func (e *PdfCpuEngine) Merge(inFiles []string, outFile string, append bool) error {
	executableFunc := api.MergeCreateFile
	if append {
		executableFunc = api.MergeAppendFile
	}
	if err := executableFunc(inFiles, outFile, false, nil); err != nil {
		return fmt.Errorf("pdfcpu merge: %w", err)
	}
	return nil
}

func (e *PdfCpuEngine) Split(inFile string, outDir string) error {
	if err := api.SplitFile(inFile, outDir, 1, e.conf); err != nil {
		return fmt.Errorf("pdfcpu split '%s' => '%s': %w", inFile, outDir, err)
	}
	return nil
}
