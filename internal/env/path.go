// internal/env/path.go

package env

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func EnsureParentDirExists(path string) error {
	dir := Dirname(path)
	if err := MkdirAll(dir); err != nil {
		return err
	}
	return nil
}

func MkdirAll(path string) error {
	return os.MkdirAll(path, 0o755)
}

func CreateTempFile(pattern string, in io.Reader) (*os.File, func(), error) {
	noop := func() {}
	// create tmp file
	tmp, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, noop, fmt.Errorf("create tmp file: %w", err)
	}
	callback := func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}
	// copy input to tmp file
	if _, err = io.Copy(tmp, in); err != nil {
		callback()
		return nil, noop, fmt.Errorf("copy to tmp file: %w", err)
	}
	// rewind to start so tmp reads from the beginning
	if _, err = tmp.Seek(0, io.SeekStart); err != nil {
		callback()
		return nil, noop, fmt.Errorf("rewind tmp file: %w", err)
	}
	return tmp, callback, nil
}

// Copy src => dst, where src and dst can be file paths or "-" for stdin/stdout.
// If dst == "", a temporary file will be created and its path returned
func CopyFile(src, dst string) error {
	var srcFile *os.File
	var dstFile *os.File
	var err error

	switch src {
	case "":
		return fmt.Errorf("copy file - source path cannot be empty")
	case "-":
		srcFile = os.Stdin
	default:
		srcFile, err = os.Open(src)
		if err != nil {
			return fmt.Errorf("copy file - open src file: %v", err)
		}
		defer srcFile.Close()
	}

	switch dst {
	case "":
		return fmt.Errorf("copy file - destination path cannot be empty")
	case "-":
		dstFile = os.Stdout
	default:
		dstFile, err = os.Create(dst)
		if err != nil {
			return fmt.Errorf("copy file - create dst file: %v", err)
		}
		defer dstFile.Close()
	}

	_, err = io.Copy(dstFile, srcFile)
	return err
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

func OpenInputFiles(path ...string) ([]*os.File, error) {
	files := make([]*os.File, len(path))
	for i, p := range path {
		switch p {
		case "":
			return nil, fmt.Errorf("input file path cannot be empty")
		case "-":
			return nil, fmt.Errorf("cannot open stdin for reading - use os.Stdin directly instead")
		}
		f, err := os.Open(p)
		if err != nil {
			// Close any files that were successfully opened
			CloseFiles(files...)
			return nil, err
		}
		files[i] = f
	}
	return files, nil
}

func OpenOutputFile(path string, append bool) (*os.File, error) {
	switch path {
	case "":
		return nil, fmt.Errorf("output cannot be empty")
	case "-":
		return nil, fmt.Errorf("cannot open stdout for writing - use os.Stdout directly instead")
	}

	var flag int = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if append {
		flag = os.O_CREATE | os.O_WRONLY | os.O_APPEND
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

func PathExists(path string) bool {
	if IsStdIO(path) {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func PathInfo(path string) (os.FileInfo, error) {
	if IsStdIO(path) {
		return nil, fmt.Errorf("cannot get file info for stdin/stdout")
	}
	return os.Stat(path)
}

func FileExists(path string) bool {
	info, err := PathInfo(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func DirExists(path string) bool {
	info, err := PathInfo(path)
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
		if path == "-" {
			return true
		}
	}
	return false
}
