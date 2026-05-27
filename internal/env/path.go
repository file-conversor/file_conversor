// internal/env/path.go

package env

import (
	"bufio"
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

// read file lines into a slice of strings (trims whitespace and ignores empty lines)
func ReadLines(path string, processLine func(string)) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line != "" {
			processLine(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan file: %w", err)
	}
	return nil
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
			return fmt.Errorf("copy file - open src file: %w", err)
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
			return fmt.Errorf("copy file - create dst file: %w", err)
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
			return nil, fmt.Errorf("open input file '%s': %w", p, err)
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
	if len(ext) == 0 {
		return true
	}
	fileExt := FileExt(path)
	for _, e := range ext {
		switch e {
		case "", "*", ".*":
			return true // skip empty extensions
		default:
			if strings.EqualFold(fileExt, e) {
				return true
			}
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
