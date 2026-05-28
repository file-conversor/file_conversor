// internal/core/output_file.go

package core

import (
	"errors"

	"github.com/file-conversor/file_conversor/internal/validation"
)

type OutputFileFlag struct {
	Overwrite    bool   // overwrite output file if it exists (default: false)
	OutputFile   string // output file path (use - for stdout)
	AcceptStdout bool   // accept stdout as output (set by command implementation)
}

func (this *OutputFileFlag) Parse(formats FormatInterface) error {
	return nil
}

func (this *OutputFileFlag) Validate(formats FormatInterface) error {
	return errors.Join(
		validation.CheckOutputStdout(this.AcceptStdout, this.OutputFile),
		validation.OutputFileExt(this.OutputFile, formats.Out()...),
		validation.OutputFileOverwritable(this.OutputFile, this.Overwrite),
	)
}
