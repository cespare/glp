package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/tools/go/vcs"
)

// This file adds some extra utility functions onto the functionality provided by go.tools/go/vcs.

type VCSCmd struct {
	*vcs.Cmd
}

type RepoRoot struct {
	VCS  VCSCmd
	Repo string
	Root string
}

func RepoRootForImportPath(importPath string, verbose bool) (*RepoRoot, error) {
	root, err := vcs.RepoRootForImportPath(importPath, verbose)
	if err != nil {
		return nil, err
	}
	return &RepoRoot{
		VCS:  VCSCmd{root.VCS},
		Repo: root.Repo,
		Root: root.Root,
	}, nil
}

func (v VCSCmd) GetRev(dir string) (rev string, dirty bool, err error) {
	switch v.Cmd.Cmd {
	case "git":
		out, err := v.runOutput(dir, "rev-parse HEAD")
		if err != nil {
			return "", false, err
		}
		rev = string(out)
		changed, err := v.runOutput(dir, "diff-index HEAD")
		if err != nil {
			return "", false, err
		}
		dirty = len(changed) > 0
	case "hg":
		out, err := v.runOutput(dir, "identify --debug --id")
		if err != nil {
			return "", false, err
		}
		outString := strings.TrimSpace(string(out))
		dirty = strings.HasSuffix(outString, "+")
		rev = strings.TrimSuffix(outString, "+")
	}
	return strings.TrimSpace(rev), dirty, nil
}

func (v VCSCmd) UpdateRev(dir, rev string) error {
	updateNeeded := false
	switch v.Cmd.Cmd {
	case "git":
		out, err := v.runSilent(dir, "cat-file -t "+rev)
		updateNeeded = (err != nil || string(bytes.TrimSpace(out)) != "commit")
	case "hg":
		panic("unimplemented")
	}
	var cmd string
	if updateNeeded {
		switch v.Cmd.Cmd {
		case "git":
			cmd = "fetch"
		case "hg":
			panic("unimplemented")
		}
		if err := v.run(dir, cmd); err != nil {
			return err
		}
	}
	switch v.Cmd.Cmd {
	case "git":
		// TODO: is it better to reset --hard instead of leaving a detached head?
		cmd = "checkout"
	case "hg":
		cmd = "update -c"
	}
	if err := v.run(dir, cmd+" "+rev); err != nil {
		return err
	}
	return nil
}

// run runs the command line cmd in the given directory. If an error occurs, run prints the command line and
// the command's combined stdout+stderr to standard error. Otherwise run discards the command's output.
func (v VCSCmd) run(dir, cmd string) error {
	_, err := v.run1(dir, cmd, true)
	return err
}

// runOutput is like run but returns the output of the command.
func (v VCSCmd) runOutput(dir, cmd string) ([]byte, error) {
	return v.run1(dir, cmd, true)
}

// runSilent is like runOutput but does not print anything on failure.
func (v VCSCmd) runSilent(dir, cmd string) ([]byte, error) {
	return v.run1(dir, cmd, false)
}

// run1 is the generalized implementation of run and runOutput.
func (v VCSCmd) run1(dir string, cmdline string, verbose bool) ([]byte, error) {
	_, err := exec.LookPath(v.Cmd.Cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"go: missing %s command. See http://golang.org/s/gogetcmd\n",
			v.Name)
		return nil, err
	}

	args := strings.Fields(cmdline)
	cmd := exec.Command(v.Cmd.Cmd, args...)
	cmd.Dir = dir
	cmd.Env = mergeEnv(os.Environ(), "PWD", cmd.Dir)
	if vcs.ShowCmd {
		fmt.Printf("cd %s\n", dir)
		fmt.Printf("%s %s\n", v.Cmd, strings.Join(args, " "))
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	out := buf.Bytes()
	if err != nil {
		if verbose || vcs.Verbose {
			fmt.Fprintf(os.Stderr, "# cd %s; %s %s\n", dir, v.Cmd, strings.Join(args, " "))
			os.Stderr.Write(out)
		}
		return nil, err
	}
	return out, nil
}
