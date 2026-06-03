// internal/core/flags/output_file_flag.go

package flags

import (
	"errors"

	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/validation"
)

type OutputFileFlag struct {
	Overwrite  bool   // overwrite output file if it exists (default: false)
	OutputFile string // output file path (use - for stdout)
	// AcceptStdout bool   // accept stdout as output (set by command implementation)
}

func (o *OutputFileFlag) Parse(formats interfaces.FormatInterface) error {
	return nil
}

func (o *OutputFileFlag) Validate(formats interfaces.FormatInterface) error {
	return errors.Join(
		validation.AllowOrDenyOutputStdout(false, o.OutputFile),
		validation.OutputFileExt(o.OutputFile, formats.Out()...),
		validation.OutputFileOverwritable(o.OutputFile, o.Overwrite),
	)
}
