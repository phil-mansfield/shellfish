# Code Outline

The purpose of this document is to give you a basic idea of how things are strucutred,
what the important data structures are, and what the general control flow through the
project is. This isn't intended to be read all at once: just focus on the things that
are relevant to what you are working on.

## What is the Execution Path of a Shellfish Program?

The workhorse of this project is the
[`Mode`](https://github.com/phil-mansfield/shellfish/blob/master/cmd/cmd.go#L29) interface.
Every different execution mode has a corresponding struct which implements the interface.
These implementations need to do three things:

1. Create an example config file.
2. Parse a config file for this mode type.
3. Execute a a mode given a particular set of config files and command line flags.

The different modes are managed by `shellfish.go`. When the command is run, the
`main()` function in this file is called. It selects to correct `Mode` implementation
and calls its methods. This file also does a lot of other boring stuff, like parsing
command line arguments, managing `stdin` and `stdout`, packing (potentially weird)
user file structures into an [`Environment`](https://github.com/phil-mansfield/shellfish/blob/master/cmd/env/env.go#L32)
struct, and handling errors that get reported by other places in the project.

Implementaitons of the `Mode` interface can be found in the [`cmd/`](https://github.com/phil-mansfield/shellfish/tree/master/cmd)
package, with each in its own file. Of the three things that `Mode`s need to do,
only the thrid one is non-trivial (the [`parse`](https://github.com/phil-mansfield/shellfish/tree/master/parse)
package takes care of the nuts-and-bolts of parsing arbitrary config files). Each mode
has a different execution path, which I will go through below:

### `id`

### `tree`

This mode is very simple. Since the only supported tree format is Behroozi et al.'s
[`consistent-trees`](https://bitbucket.org/pbehroozi/consistent-trees),
this file just calls functions from the
[`los/tree`](https://github.com/phil-mansfield/shellfish/tree/master/los/tree)
package, which are themselves wrappers around `consistent-trees`'s API.

### coord

### shell

### stats

### prof

## What Files are Difficult to Read?

- `cmd/stats.go`
- `cmd/prof.go`
- `cmd/memo/memoize.go`
- `los/analyze/kde.go`

And to a lesser extent,

- `cmd/shell.go`
