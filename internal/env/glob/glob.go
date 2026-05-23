// internal/env/glob/glob.go

package glob

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/file-conversor/file_conversor/internal/env"
)

type Globber struct {
	Exts  []string
	regex *regexp.Regexp
}

func New(exts ...string) (*Globber, error) {
	newExts := make([]string, len(exts))
	copy(newExts, exts)

	if len(newExts) == 0 {
		newExts = append(newExts, ".*") // if no extensions provided, match all files
	}
	for i, _ := range newExts {
		newExts[i] = strings.TrimPrefix(newExts[i], ".")
	}
	pattern := fmt.Sprintf(`\.(%s)$`, strings.Join(newExts, "|"))
	regexComp, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
	}
	return &Globber{
		Exts:  newExts,
		regex: regexComp,
	}, nil
}

func (g *Globber) GlobFiles(roots ...string) ([]string, error) {
	var files = make([]string, 0, 20)

	for _, input := range roots {
		if g.regex.MatchString(input) {
			files = append(files, input)
		} else if env.FileExt(input) == "" {
			// Find all matching files in the directory
			err := g.RecurseDirectory(func(filePath string) error {
				files = append(files, filePath)
				return nil
			}, input)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("input '%s' does not match pattern '%s'", input, g.regex.String())
		}
	}
	return files, nil
}

func (g *Globber) RecurseDirectory(parseFile func(string) error, paths ...string) error {
	for _, path := range paths {
		err := filepath.WalkDir(path, func(p string, info os.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("recurse dir open path: %w", err)
			}
			if !info.IsDir() && g.regex.MatchString(info.Name()) {
				return parseFile(p)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("recurse dir failed for '%s': %w", path, err)
		}
	}
	return nil
}
