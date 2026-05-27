// internal/env/tmp.go

package env

import (
	"fmt"
	"io"
	"os"
)

func CopyToTmpFileRaw(pattern string, in io.Reader) (*os.File, func(), error) {
	noop := func() {}
	// create tmp file
	if pattern == "" {
		pattern = "tmpfile-*"
	}
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

// Copies the file at the given path to a temporary file.
// Returns the temporary file and a cleanup function.
func CopyToTmpFile(path string) (*os.File, func(), error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("copy to tmp - open file: %w", err)
	}
	defer file.Close()

	tmpFile, cleanup, err := CopyToTmpFileRaw("", file)
	if err != nil {
		return nil, nil, fmt.Errorf("copy to tmp - copy %s to tmp file: %w", path, err)
	}
	return tmpFile, cleanup, nil
}
