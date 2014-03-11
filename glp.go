package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	projectDirName = "glp"
	cacheDirName   = "_cache"
	pinlistName    = "deps.json"
)

var goBinary = ""

func init() {
	var err error
	goBinary, err = exec.LookPath("go")
	if err != nil {
		fatal("No go tool executable located -- is Go installed?")
	}
}

// findProjectRoot locates a glp project root by looking for a 'glp' directory.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		stat, err := os.Stat(filepath.Join(dir, projectDirName))
		if err == nil && stat.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no glp project directory found")
		}
		dir = parent
	}
}

// makeGOPATH constructs a $GOPATH from an absolute project root directory.
func makeGOPATH(root string) string {
	cacheDir := filepath.Join(root, projectDirName, cacheDirName)
	return fmt.Sprintf("%s:%s", root, cacheDir)
}

// mergeEnv merges a key/value pair into an environment variable list, replacing any existing matching keys.
func mergeEnv(env []string, key, value string) []string {
	result := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, key+"=") {
			result = append(result, e)
		}
	}
	return append(result, key+"="+value)
}

func execGoTool(args []string, env []string) error {
	args = append([]string{"go"}, args...)
	return syscall.Exec(goBinary, args, env)
}

var disabledGoCommands = map[string]bool{
	"install": true,
	"get":     true,
}

func main() {
	args := os.Args[1:]

	root, err := findProjectRoot()
	if err != nil {
		fatalf("Error locating glp project: %s", err)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		fatal(err)
	}

	gopath := makeGOPATH(root)

	if len(args) > 0 {
		command := args[0]
		switch command {
		case "sync":
			if err := Sync(root, gopath); err != nil {
				fatal(err)
			}
			return
		default:
			if disabledGoCommands[command] {
				fatalf("Error: the command 'go %s' cannot be used in a glp project.\n", command)
			}
		}
	}

	// Read the deps file
	pinlistFilename := filepath.Join(root, projectDirName, pinlistName)
	pinlist, err := LoadPinlist(pinlistFilename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("No pinlist file (looked for %s). Try running 'glp sync'.\n", pinlistFilename)
			os.Exit(0)
		}
		fatal(err)
	}

	if err := Verify(root, pinlist); err != nil {
		fatal(err)
	}

	// Write out the pinlist so it's normalized in case it has been hand-edited.
	if err := pinlist.Save(pinlistFilename); err != nil {
		fatal(err)
	}

	// Pass through with a modified GOPATH.
	env := mergeEnv(os.Environ(), "GOPATH", gopath)
	fatal(execGoTool(args, env))
}

func fatal(args ...interface{}) {
	fmt.Println(args...)
	os.Exit(1)
}

func fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}
