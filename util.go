package main

import (
	"os"
	"path/filepath"
)

var ignoreDirs = map[string]bool{
	".git": true,
	".hg":  true,
}

// FindDirsRecursively finds all directories in path, recursively (but skipping VCS metadata directories).
func FindDirsRecursively(path string) []string {
	var results []string
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if ignoreDirs[filepath.Base(path)] {
				return filepath.SkipDir
			}
			results = append(results, path)
		}
		return nil
	})
	return results
}
