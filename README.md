## glp

glp is a tool for managing [Go](http://golang.org/) code and dependencies in a project-oriented fashion. It is
intended for use with applications (not libraries).

glp is the evolution of [go-localpath](https://github.com/cespare/go-localpath).

Before using glp, it is recommended that you have a good grasp of the usual, global-`$GOPATH`-oriented way of
organizing Go code and libraries, and that you have a working Go installation that follows this model. Read
[How to Write Go Code](http://golang.org/doc/code.html) for a good introduction.

**Note:** glp is still a work in progress and may change at any time.

### Installation

You'll need a normal Go development environment set up first.

    $ go get github.com/cespare/glp

### Usage

glp projects are denoted by the presence of a directory called `glp/` in the project root. If this does not
exist, glp simply `exec`s `go` directly, transparently passing through the arguments and environment.

If glp is inside a glp project (in a glp project root or anywhere below a glp project root), then glp reads
the *pinlist* at `glp/deps.json` to see what dependencies and versions are specified. After ensuring that the
local dependency package cache (`glp/_cache`) matches the versions in the pinlist, `glp` calls `go` with a
modified `$GOPATH`:

    GOPATH=/path/to/project:/path/to/project/glp/_cache

glp has one special command: `sync`. It takes no arguments and synchronizes the state of the dependencies in
project code code (any Go packages found in the project root or `src/`), the dependencies in `glp/deps.json`,
and the dependency cache. The canonical list of dependencies is taken to be that needed by the project code.
Versions are taken from the current pinlist. Any dependencies not currently in the pinlist are downloaded and
the latest version is used. The cache is modified to reflect the pinned versions and the pinlist is updated
with any missing dependencies.

This means that starting a new project is as simple as making an empty `glp` directory and running `glp sync`.
The pinlist and cache will be created from scratch (using the latest versions of the dependencies).

Because of the existence of `glp sync`. glp disables `go get` and `go install` (their behavior would be
confusing if allowed).

TODO: more thorough description. Consider splitting out a quick-start guide and a detailed and complete
explanation of glp's behavior.

### Tips

* You should configure your VCS to ignore the `glp/_cache` directory (but leave `glp/` and any other files
  inside it alone). If you're using git, you could put `/glp/_cache` in the `.gitignore` in your project root.

### Motivation

Here are some of the goals behind glp:

* Pin dependencies to particular revisions without having to check the code into the repository or use
  submodules.
* Build code in an isolated environment that cannot accidentally pull in dependencies from the global
  `$GOPATH`.
* Provide tooling for discovering and downloading/pinning new dependencies.
* No switching cost when working on a project (such as sourcing setup shell scripts).
* Keep application code anywhere (not in a predefined hierarchy determined by a global `$GOPATH`).
* Support all the version control tools that the Go tool supports (except svn).
* Allow for keeping root both in the project root or split into packages in the `src` directory.
* For projects with multiple packages, import the package like `"foo"`, not `"github.com/org/proj/foo"`.

### Similar projects

* [godep](https://github.com/kr/godep)
* [Johnny Deps](https://github.com/VividCortex/johnny-deps)
* [gpm](https://github.com/pote/gpm)/[gvp](https://github.com/pote/gvp)
* [rx](http://godoc.org/kylelemons.net/go/rx)
* [so many others](https://docs.google.com/a/liftoff.io/document/d/1k-3mwBqAdTIKGcilWZPuKSMy3DWtfNRFDs9o98lcwHY/edit#)
