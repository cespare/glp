package main

const glpHelp = `glp -- dependency pinning for Go
Usage:
    %s COMMAND [OPTIONS]...
glp provides several builtin commands. All other subcommands are delegated to
the go tool using a modified $GOPATH for the glp project.

glp commands:

help
    Show this help.
path
    Print the GOPATH with which glp calls the Go tool.
sync
    Synchronize the project dependencies (from the source), the pinned
    versions (in glp/deps.json), and the cache (glp/_cache).

For more information, see https://github.com/cespare/glp.
`
