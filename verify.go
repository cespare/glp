package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type mismatchedDepsError struct {
	error
}

// Verify checks a project's cache directory against the dep list and returns an error if they do not match.
func Verify(root string, pinlist *Pinlist) error {
	for _, dep := range pinlist.Deps {
		if err := verify(root, dep); err != nil {
			return err
		}
	}
	return nil
}

func verify(root string, dep Dep) error {
	repo, err := RepoRootForImportPath(dep.Name, false)
	if err != nil {
		return err
	}
	dir := filepath.Join(root, projectDirName, cacheDirName, "src", repo.Root)
	rev, dirty, err := repo.VCS.GetRev(dir)
	if err != nil {
		return err
	}
	if rev != dep.Rev {
		err := fmt.Errorf("Pinlist has version %s for %s, but found %s in cache.", dep.Rev, dep.Name, rev)
		return mismatchedDepsError{err}
	}
	if dirty {
		fmt.Fprintf(os.Stderr, "Warning: found dirty repo for dependency %s\n", dep.Name)
	}
	return nil
}
