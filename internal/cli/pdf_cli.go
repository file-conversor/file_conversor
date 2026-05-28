// internal/cli/pdf/pdf_cli.go

package cli

type PdfCLI struct {
	Decrypt PdfDecryptCLI `cmd:""  help:"Decrypt PDF files with a password."`
	Encrypt PdfEncryptCLI `cmd:""  help:"Encrypt PDF files with a password and permissions."`
	Merge   PdfMergeCLI   `cmd:""  help:"Merge PDF files into a single PDF."`
}
