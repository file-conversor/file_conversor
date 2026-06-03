// internal/core/flags/output_dir_flag.go

package flags

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/file-conversor/file_conversor/internal/env"
	"github.com/file-conversor/file_conversor/internal/interfaces"
	"github.com/file-conversor/file_conversor/internal/utils"
	"github.com/file-conversor/file_conversor/internal/validation"
)

type OutputDirFlag struct {
	Overwrite bool   // overwrite output file if it exists (default: false)
	OutputDir string // output dir path (default: current dir)
	Suffix    string // suffix to add to output file stem (default: "")
	Format    string // output file format (e.g. ".pdf")
	// AcceptStdout bool   // whether to accept stdout as output dir
}

func (o *OutputDirFlag) Parse(formats interfaces.FormatInterface) error {
	return errors.Join(
		validation.OutputDirEnsure(o.OutputDir),
	)
}

func (o *OutputDirFlag) Validate(formats interfaces.FormatInterface) error {
	return errors.Join(
		validation.AllowOrDenyOutputStdout(false, o.OutputDir),
		validation.OutputFileExt(o.Format, formats.Out()...),
		validation.OutputValidSuffix(o.Suffix),
		validation.OutputDirExists(o.OutputDir),
	)
}

func (o *OutputDirFlag) GetOutputPath(inputPath string) (outPath string, errGrp error) {
	var stemStrBuilder strings.Builder
	stemStrBuilder.Grow(64)

	if err := errors.Join(errGrp,
		utils.ExtractError(stemStrBuilder.WriteString(env.FileStem(inputPath))),
		utils.ExtractError(stemStrBuilder.WriteString(o.Suffix)),
		utils.ExtractError(stemStrBuilder.WriteString(o.Format)),
	); err != nil {
		errGrp = fmt.Errorf("input '%s' - failed output build: %w", inputPath, err)
		return
	}
	outPath = filepath.Join(o.OutputDir, stemStrBuilder.String())

	if err := errors.Join(
		validation.InputOutputNotEqual(outPath, inputPath),
		validation.OutputFileOverwritable(outPath, o.Overwrite),
	); err != nil {
		errGrp = errors.Join(errGrp, fmt.Errorf("input '%s' - output '%s': %w", inputPath, outPath, err))
	}
	return
}

func (o *OutputDirFlag) GetOutputFile(inputPath string) (outFile *os.File, errGrp error) {
	// get output file path for input file
	outFilePath, err := o.GetOutputPath(inputPath)
	if err != nil {
		errGrp = fmt.Errorf("get output file path: %w", err)
		return
	}

	// open output file
	outFile, err = os.Create(outFilePath)
	if err != nil {
		errGrp = fmt.Errorf("open output file '%s': %w", outFilePath, err)
		return
	}

	return outFile, nil
}
