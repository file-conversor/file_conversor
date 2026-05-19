// internal/env/file.go

package env

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

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

func EnsureParentDirExists(path string) error {
	dir := Dirname(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return nil
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
