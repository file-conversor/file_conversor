// internal/env/file.go

package env

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func EnsureParentDirExists(path string) error {
	dir := Dirname(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func ToIoReadSeeker(files ...*os.File) ([]io.ReadSeeker, error) {
	ioReadSeekers := make([]io.ReadSeeker, len(files))
	for i, f := range files {
		if f == nil {
			return nil, errors.New("nil file pointer cannot be converted to io.ReadSeeker")
		}
		ioReadSeekers[i] = f
	}
	return ioReadSeekers, nil
}

func CloseFiles(res ...*os.File) error {
	var errGrp error
	for _, f := range res {
		if f == nil {
			continue // skip nil pointers
		}
		if err := (*f).Close(); err != nil {
			errGrp = errors.Join(errGrp, err)
		}
	}
	return errGrp
}

func OpenInputFiles(path ...string) ([]*os.File, []io.ReadSeeker, error) {
	files := make([]*os.File, len(path))
	for i, p := range path {
		f, err := os.Open(p)
		if err != nil {
			// Close any files that were successfully opened
			CloseFiles(files...)
			return nil, nil, err
		}
		files[i] = f
	}
	ioReadSeekers, err := ToIoReadSeeker(files...)
	if err != nil {
		CloseFiles(files...)
		return nil, nil, err
	}
	return files, ioReadSeekers, nil
}

func OpenOutputFile(path string, append bool) (*os.File, error) {
	if err := EnsureParentDirExists(path); err != nil {
		return nil, err
	}

	var flag int = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if append {
		flag = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	}
	f, err := os.OpenFile(path, flag, 0o644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func GetOutputFile(input, outputDir, outputSuffix, outputExt string) string {
	var stemStrBuilder strings.Builder
	stemStrBuilder.Grow(64)

	stemStrBuilder.WriteString(FileStem(input))
	stemStrBuilder.WriteString(outputSuffix)
	stemStrBuilder.WriteString(outputExt)

	return filepath.Join(outputDir, stemStrBuilder.String())
}

func ChangeFileExt(path, newExt string) string {
	return path[:len(path)-len(FileExt(path))] + newExt
}

func ChangeFileStem(path, newStem string) string {
	return filepath.Join(Dirname(path), newStem+FileExt(path))
}

func Dirname(path string) string {
	return filepath.Dir(path)
}

func BaseName(path string) string {
	return filepath.Base(path)
}

func FileStem(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

func FileExt(path string) string {
	return filepath.Ext(path)
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func IsEqualPath(path1, path2 string) (bool, error) {
	abs1, err1 := filepath.Abs(path1)
	abs2, err2 := filepath.Abs(path2)
	if err1 != nil || err2 != nil {
		return false, errors.Join(err1, err2)
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(abs1, abs2), nil
	}
	return abs1 == abs2, nil
}

func IsFileExt(path string, ext ...string) bool {
	fileExt := FileExt(path)
	for _, e := range ext {
		if strings.EqualFold(fileExt, e) {
			return true
		}
	}
	return false
}

// checks if is stdin or stdout
func IsStdIO(paths ...string) bool {
	for _, path := range paths {
		if path != "-" {
			return false
		}
	}
	return true
}
