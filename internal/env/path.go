// internal/env/path.go

package env

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

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
