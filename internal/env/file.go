// internal/env/file.go

package env

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
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

func OpenInputFiles(callback func(file *os.File) error, inputFiles ...string) (cleanupFuncs []func() error, errGrp error) {
	var readStdin bool

	for _, input := range inputFiles {
		switch input {
		case "":
			errGrp = errors.Join(errGrp, fmt.Errorf("input file path cannot be empty"))
		case "-":
			if readStdin {
				continue // already reading from stdin, skip additional "-"
			}
			readStdin = true
			tmpFile, cleanup, err := CopyToTmpFileRaw("", os.Stdin)
			if err != nil {
				errGrp = errors.Join(errGrp, fmt.Errorf("copy stdin to tmp file: %w", err))
			} else {
				// ensure tmp file is cleaned up
				cleanupFuncs = append(cleanupFuncs, func() error { cleanup(); return nil })
				// pass tmp file to callback for processing
				callback(tmpFile)
			}
		default:
			inFile, err := os.Open(input)
			if err != nil {
				errGrp = errors.Join(errGrp, fmt.Errorf("open input file '%s': %w", input, err))
			} else {
				cleanupFuncs = append(cleanupFuncs, inFile.Close) // ensure file is closed
				callback(inFile)
			}
		}
		if errGrp != nil {
			for _, cleanup := range cleanupFuncs {
				cleanup() // clean up any files that were opened before returning error
			}
			cleanupFuncs = []func() error{}
			return
		}
	}
	return
}

func OpenOutputFile(callback func(file *os.File) error, outputFile string) (cleanup []func() error, errGrp error) {
	switch outputFile {
	case "":
		errGrp = errors.Join(errGrp, fmt.Errorf("output cannot be empty"))
	case "-":
		callback(os.Stdout) // pass stdout file handle to callback for writing output
	default:
		outFile, err := os.Create(outputFile)
		if err != nil {
			errGrp = errors.Join(errGrp, fmt.Errorf("open output file '%s': %w", outputFile, err))
		} else {
			// ensure file is closed
			cleanup = append(cleanup, outFile.Close)
			// pass output file handle to callback for writing output
			callback(outFile)
		}
	}
	if errGrp != nil {
		for _, cleanupFunc := range cleanup {
			cleanupFunc() // clean up any files that were opened before returning error
		}
		cleanup = []func() error{}
	}
	return
}
