// internal/cli/flags/pdf_encryption.go

package flags

import "github.com/file-conversor/file_conversor/internal/core/pdf"

type PdfEncryptionFlag struct {
	Encryption string `short:"e" optional:"" default:"aes256" enum:"aes256,aes128,rc128,rc40" help:"Encryption algorithm and key size (${enum})."`
}

func (c *PdfEncryptionFlag) Get() pdf.EncryptionAlgorithm {
	switch c.Encryption {
	case "rc40":
		return pdf.RC40
	case "rc128":
		return pdf.RC128
	case "aes128":
		return pdf.AES128
	default:
		return pdf.AES256
	}
}
