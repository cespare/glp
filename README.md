# glp

glp is a tool for managing [Go](http://golang.org/) code and dependencies in a project-oriented fashion. It is
intended for use with applications (not libraries).

glp is the evolution of [go-localpath](https://github.com/cespare/go-localpath).

Before using glp, it is recommended that you have a good grasp of the usual, global-`$GOPATH`-oriented way of
organizing Go code and libraries, and that you have a working Go installation that follows this model. Read
[How to Write Go Code](http://golang.org/doc/code.html) for a good introduction.

**Note:** glp is still a work in progress and may change at any time.

## Installation

You'll need a normal Go development environment set up first.

    $ go get github.com/cespare/glp

## Quick start

If you have a project that already uses glp, you can fetch the dependencies by running `glp sync` from
somewhere underneath the project root. After this, you can use `glp build`, `glp test`, and any other command
the go tool offers (with a couple of exceptions; see "Disabled commands" below).

### New project setup

If you have a new project that resides in a single Go package, setting it up to use glp is very easy. (See
"Project Layout", below, if you have multiple packages). The code doesn't have to reside in any particular
place (unlike the typical Go project configuration, it does not have to be within your $GOPATH). In the
project root (that is, the directory containing your `package main` files), just do this:

    $ mkdir glp
    $ glp sync

The sync step will discover all dependencies your code has, download the latest versions into a local cache,
and add the versions to `glp/deps.json`. You should configure your vcs to ignore `glp/_cache` and commit
`glp/deps.json`.

## Usage

glp projects are denoted by the presence of a directory called `glp/` in the project root. If this does not
exist (in the current directory or any ancestor in the directory tree), glp exits with an error message.

First, glp reads the *pinlist* at `glp/deps.json` to see what dependencies and versions are specified. After
ensuring that the local dependency package cache (`glp/_cache`) matches the versions in the pinlist, `glp`
calls `go` with a modified `$GOPATH`:

    GOPATH=/path/to/project:/path/to/project/glp/_cache

glp has some special commands which are interpreted directly. These are each listed below.

### sync

`glp sync` takes no arguments. It synchronizes the state of the dependencies in project code (any Go packages
found in the project root or `src/`), the dependencies in `glp/deps.json`, and the dependency cache.

The canonical list of dependencies is taken to be that needed by the project code. Versions are taken from the
current pinlist. Any dependencies not currently in the pinlist are downloaded and the latest version is used.
The cache is modified to reflect the pinned versions and the pinlist is updated with any missing dependencies.

### path

`glp path` prints out the `$GOPATH` that glp uses when it invokes the Go tool. This can be useful, for example,
if you're modifying some other tool which calls `go` to work with glp.

### help

`glp help` prints out a help message with some basic usage instructions.

### Updating dependencies

If you want to update the version of a dependency in use, there are two ways to do it.

If you want to update `github.com/foo/bar` to the latest revision:

* Delete the package from the cache: `rm -rf glp/_cache/src/github.com/foo/bar`
* Edit `glp/deps.json` and delete the entry for `github.com/foo/bar`
* Run `glp sync` -- this pulls down the latest version to the cache and adds it to `deps.json`

If you want to update to a specific revision:

* Edit `glp/deps.json` and change the `"rev"` field to be the new revision ID
* Run `glp sync` to update the cache

There is [an issue](https://github.com/cespare/glp/issues/8) to add a command that does these steps for you.

### Disabled go commands

Because of the existence of `glp sync`, `glp get` and `glp install` are disabled (their behavior would be
confusing if allowed, and the glp workflow replaces their functionality).

## Tips

* You should configure your VCS to ignore the `glp/_cache` directory (but leave `glp/` and any other files
  inside it alone). If you're using git, you could put `/glp/_cache` in the `.gitignore` in your project root.

## Motivation

Here are some of the goals behind glp:

* Pin dependencies to particular revisions without having to check the code into the repository or use
  submodules.
* Build code in an isolated environment that cannot accidentally pull in dependencies from the global
  `$GOPATH`.
* Provide tooling for discovering and downloading/pinning new dependencies.
* No switching cost when working on a project (such as sourcing setup shell scripts).
* Keep application code anywhere (not in a predefined hierarchy determined by a global `$GOPATH`).
* Allow for keeping root both in the project root or split into packages in the `src` directory.
* For projects with multiple packages, import the package like `"foo"`, not `"github.com/org/proj/foo"`.

## Major To-Dos

* [Mercurial support](https://github.com/cespare/glp/issues/6)
* [`glp update` command](https://github.com/cespare/glp/issues/8)

## Similar projects

* [godep](https://github.com/kr/godep)
* [Johnny Deps](https://github.com/VividCortex/johnny-deps)
* [gpm](https://github.com/pote/gpm)/[gvp](https://github.com/pote/gvp)
* [rx](http://godoc.org/kylelemons.net/go/rx)
* [so many others](https://code.google.com/p/go-wiki/wiki/PackageManagementTools)
