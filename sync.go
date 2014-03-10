package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
)

// Sync synchronizes the state of the project code (what dependencies it has), the pinlist, and all
// (transitive) dependencies in the cache.
//   * First, analyze the project code to get a list of its dependencies
//   * For each dependency, update the cache/dependency list:
//     - If the dependency is not cached, pull down the rev specifed (or latest if dep not in the pinlist)
//     - If the dependency is cached and the cached repo is dirty, fail
//     - If the dependency is cached and the cached rev is different from the specified rev, update the cached
//			 repo to the specified rev
//		 - Add the dep's transitive non-stdlib deps to the dependency list to be processed (if not done already)
//		* Recreate the pinlist from the set of deps and versions in the updated dependency list (this will also
//			have the effect of removing outdated dependencies)
//		* Write out the new pinlist
func Sync(root, gopath string) error {
	// Make a map out of the current pinlist
	pinnedVersions, err := makePinMap(root)
	if err != nil {
		return err
	}

	// Get the list of immediate dependencies from the code
	context := new(build.Context)
	*context = build.Default
	context.GOPATH = gopath

	immediateDeps, projectPackages, err := findProjectDeps(context, root)
	if err != nil {
		return err
	}

	fmt.Println("Found project packages:")
	for _, pkg := range projectPackages {
		fmt.Printf("\t%s\n", pkg)
	}

	// Now sync each dependency, adding transitive deps to the to-process list as we go.
	var processedDeps smap
	toProcessDeps := immediateDeps
	// Multiple packages may map to the same repo, and thus must necessarily must be at the same rev. We will
	// fill in the following two mappings as we go and fail if there's a contradiction.
	pinnedVersionsByRepo := make(map[string]string)
	packageToRepo := make(map[string]string)
	cacheDir := filepath.Join(root, projectDirName, cacheDirName, "src")
	for len(toProcessDeps) > 0 {
		importPath := toProcessDeps[0]
		toProcessDeps.Remove(importPath)
		repo, err := RepoRootForImportPath(importPath, false)
		if err != nil {
			return err
		}
		switch repo.VCS.Cmd.Cmd {
		case "git", "hg":
		default:
			return fmt.Errorf("%s is not VCS supported by glp", repo.VCS.Cmd.Name)
		}
		packageToRepo[importPath] = repo.Root
		rev, pinned := pinnedVersions[importPath]
		if pinned {
			rev2, ok := pinnedVersionsByRepo[repo.Root]
			if ok && rev != rev2 {
				return fmt.Errorf("Multiple packages with conflicting revs map to the repo %s", repo.Root)
			}
		}

		// Check if the repo is already cached at all.
		repoDir := filepath.Join(cacheDir, repo.Root)
		if _, err := os.Stat(repoDir); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			// Repo hasn't been downloaded; fetch latest
			fmt.Printf("Dep repo %s does not exist; downloading...", repo.Root)
			if err := repo.VCS.Create(repoDir, repo.Repo); err != nil {
				return err
			}
			fmt.Println("done.")
		}

		// Get the state of the repo (current rev and whether it's dirty)
		currentRev, dirty, err := repo.VCS.GetRev(repoDir)
		if err != nil {
			return err
		}
		if dirty {
			return fmt.Errorf("dep repo at %s is currently dirty (has untracked changes)", repoDir)
		}
		if pinned && currentRev != rev {
			// If the repo is does not matched the pinned version, update to that version
			fmt.Printf("Updating dep repo at %s to rev %s\n", repoDir, rev)
			if err := repo.VCS.UpdateRev(repoDir, rev); err != nil {
				return err
			}
		}
		if !pinned {
			rev = currentRev
		}

		// Update the pin list -- these are possibly no-ops
		pinnedVersions[importPath] = rev
		pinnedVersionsByRepo[repo.Root] = rev

		// Add the dependencies of this package to the to-process list if they haven't already been inspected
		deps, err := findDeps(context, repoDir)
		if err != nil {
			return err
		}
		for _, dep := range deps {
			if !processedDeps.Contains(dep) {
				toProcessDeps.Add(dep)
			}
		}

		processedDeps.Add(importPath)
	}
	fmt.Println("Dependencies are up-to-date.")

	// Now write out the updated pin list. We reconstruct from pinnedVersionsByRepo and packageToRepo becaused
	// pinnedVersions may contain outdated deps that are not needed by the current project code.
	pinlistFilename := filepath.Join(root, projectDirName, pinlistName)
	newPinlist := new(Pinlist)
	var deps smap
	for dep := range packageToRepo {
		deps.Add(dep)
	}
	for _, dep := range deps {
		newPinlist.Deps = append(newPinlist.Deps, Dep{
			Name: dep,
			Rev:  pinnedVersionsByRepo[packageToRepo[dep]],
		})
	}
	return newPinlist.Save(pinlistFilename)
}

// makePinMap creates a map of name -> rev from the entries in the pinlist in the project located at root.
func makePinMap(root string) (map[string]string, error) {
	pinnedVersions := make(map[string]string)
	pinlistFilename := filepath.Join(root, projectDirName, pinlistName)
	pinlist, err := LoadPinlist(pinlistFilename)
	switch {
	case err == nil:
		for _, dep := range pinlist.Deps {
			pinnedVersions[dep.Name] = dep.Rev
		}
	case os.IsNotExist(err):
	default:
		return nil, err
	}
	return pinnedVersions, nil
}

func findProjectDeps(context *build.Context, root string) (immediateDeps, projectPackages smap, err error) {
	srcDir := filepath.Join(root, "src")
	possiblePackageDirs := FindDirsRecursively(srcDir)
	possiblePackageDirs = append(possiblePackageDirs, ".")
	for _, dir := range possiblePackageDirs {
		pkg, err := context.ImportDir(dir, 0)
		if err != nil {
			if _, ok := err.(*build.NoGoError); ok {
				continue
			}
			return nil, nil, err
		}

		packageName := dir
		if dir != "." {
			packageName, err = filepath.Rel(srcDir, dir)
			if err != nil {
				return nil, nil, err
			}
		}
		projectPackages.Add(packageName)
		for _, dep := range findNonStdDeps(context, pkg) {
			immediateDeps.Add(dep)
		}
	}

	// Remove the intra-project deps from immediateDeps
	for _, dep := range immediateDeps {
		if projectPackages.Contains(dep) {
			immediateDeps.Remove(dep)
		}
	}

	return immediateDeps, projectPackages, nil
}

// Find the first-level deps for a package
func findDeps(context *build.Context, dir string) (immediateDeps smap, err error) {
	pkg, err := context.ImportDir(dir, 0)
	if err != nil {
		return nil, err
	}
	return findNonStdDeps(context, pkg), nil
}

func findNonStdDeps(context *build.Context, pkg *build.Package) (immediateDeps smap) {
	imports := pkg.Imports
	imports = append(imports, pkg.TestImports...)
	imports = append(imports, pkg.XTestImports...)
	for _, imp := range imports {
		if !pkgIsStd(context, imp) {
			immediateDeps.Add(imp)
		}
	}
	return immediateDeps
}

func pkgIsStd(context *build.Context, importPath string) bool {
	pkg, err := context.Import(importPath, ".", 0)
	return err == nil && pkg.Goroot
}
