// internal/core/output_file.go

package core

import (
	"errors"

	"github.com/file-conversor/file_conversor/internal/validation"
)

type OutputFlag struct {
	Overwrite  bool   // overwrite output file if it exists (default: false)
	OutputFile string // output file path (use - for stdout)
}

func (this *OutputFlag) Parse(formats Format) error {
	return nil
}

func (this *OutputFlag) Validate(acceptStdout bool, formats Format) error {
	return errors.Join(
		validation.CheckOutputStdout(acceptStdout, this.OutputFile),
		validation.OutputFileExt(this.OutputFile, formats.Out()...),
		validation.OutputFileOverwritable(this.OutputFile, this.Overwrite),
	)
}
