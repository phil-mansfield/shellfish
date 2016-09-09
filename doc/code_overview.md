# Code Overview

### Links to Resources

- [Documentation Page](https://godoc.org/github.com/phil-mansfield/shellfish)
- [Quick Descriptions of Project, Package and File Structure](https://github.com/phil-mansfield/shellfish/blob/master/doc/outline.md)
- [Go tutorial](tour.golang.org)

### Language

Shellfish is written in the programming lagnuage [Go](golang.org). Go is very similar
to C: if you are already familiar with C, you can learn the important parts of Go in
an afternoon. The most important differences are that Go is garbage collected, memory-safe,
and allows for simple methods to be attached to structs. (Go is kind of trendy among certain
types of software engineers because of some its more advanced features, but I keep these
features to a minimum in Shellfish specfically to make it easier to contribute.)

You can learn the basics of the language [here](tour.golang.org). If you write out the
examples, you'll be done in an hour or two. Don't worry about the finer points of
concurrency or interfaces: you can skip those sections.

### Asking Questions

The best way to ask questions is through the Issues tab near the top of the page. 

### Contributing Changes

The simplest way to contribute changes to the Shellfish project is through github.
First, create a _fork_ of the project using the project in the upper right 

### Contribution guidelines



## Language/Project Quirks

### Debugging Print Statements

Because Shellfish uses `stdout` to communicate between different processes, you will want
to make sure that all your print statements print to `stderr`. The simplest way to do this
is with the `log` package. At the top of a file, add `"log"` to the import list and use
`log.Printf(...)`

### Profiling

Go supports 
