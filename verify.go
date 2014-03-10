package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
)

type mismatchedDepsError struct {
	error
}

// Verify checks a project's cache directory against the dep list and returns an error if they do not match.
func Verify(root string, pinlist *Pinlist) error {
	// Find all the repos containing Go packages in the cache.
	cache := filepath.Join(root, projectDirName, cacheDirName, "src")
	context := new(build.Context)
	*context = build.Default
	context.GOPATH = cache
	var cachedRepos smap
	for _, path := range FindDirsRecursively(cache) {
		_, err := context.ImportDir(path, 0)
		if err != nil {
			if _, ok := err.(*build.NoGoError); ok {
				continue
			}
			return err
		}
		importPath, err := filepath.Rel(cache, path)
		if err != nil {
			return err
		}
		repoRoot, err := RepoRootForImportPath(importPath, false)
		if err != nil {
			return err
		}
		cachedRepos.Add(repoRoot.Root)
	}

	// Verify that the package exists with the correct version for each package in the pinlist. Remove from the
	// list of cached repos as we go.
	for _, dep := range pinlist.Deps {
		if err := verify(root, dep); err != nil {
			return err
		}
		repoRoot, err := RepoRootForImportPath(dep.Name, false)
		if err != nil {
			return err
		}
		cachedRepos.Remove(repoRoot.Root)
	}

	// If there are any repos left in cachedRepos, they are not used by any packages in the pinlist. We need to
	// remove these to avoid polluting the namespace and allowing untracked dependencies in the build.
	for _, repo := range cachedRepos {
		fmt.Printf("Removing cached repo not in the pinlist: %s\n", repo)
		if err := os.RemoveAll(filepath.Join(cache, repo)); err != nil {
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
	// Check if the directory exists first just to provide a friendlier warning.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("Repo %s (for dependency %s) not cached. Run 'glp sync'.", repo.Repo, dep.Name)
	}
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
