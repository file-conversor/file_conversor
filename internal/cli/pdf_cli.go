// internal/cli/pdf/pdf_cli.go

package cli

type PdfCLI struct {
	Merge PdfMergeCLI `cmd:""  help:"Merge PDF files into a single PDF."`
}
