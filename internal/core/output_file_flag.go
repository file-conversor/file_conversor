// internal/core/output_file_flag.go

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

func (o *OutputFileFlag) Parse(formats FormatInterface) error {
	return nil
}

func (o *OutputFileFlag) Validate(formats FormatInterface) error {
	return errors.Join(
		validation.AllowOrDenyOutputStdout(o.AcceptStdout, o.OutputFile),
		validation.OutputFileExt(o.OutputFile, formats.Out()...),
		validation.OutputFileOverwritable(o.OutputFile, o.Overwrite),
	)
}
