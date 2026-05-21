// internal/pdf/merge.go

package pdf

import (
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Merge merges multiple PDF files into a single PDF file.
func Merge(append bool, output string, inputs ...string) error {
	if output == "-" {
		return api.Merge("", inputs, os.Stdout, nil, false)
	}

	if append {
		return api.MergeAppendFile(inputs, output, false, nil)
	}
	return api.MergeCreateFile(inputs, output, false, nil)
}
