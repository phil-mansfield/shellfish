# Code Overview

### Links to Resources

- [Documentation Page](https://godoc.org/github.com/phil-mansfield/shellfish)
- [Quick Descriptions of Project, Package and File Structure](https://github.com/phil-mansfield/shellfish/blob/master/doc/outline.md)
- [Go tutorial](tour.golang.org)
- [Contributing to Open Source](https://guides.github.com/activities/contributing-to-open-source/)

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

The best way to ask questions is through the Issues tab near the top of the page. Once
in that tab, click the green "New Issue" button and write out your question. When you
want to submit, first select "question" from the drop down menu and then press the green
"Submit New Issue" button.

The benefits of using the Issues pages is that is unifies all the questions, bug reports,
and feature requests in one place for me, allows me to reference back to common questions,
and it allows all team members to deal with problems. You will need a github account, but
those are quick to make.

### Contributing Changes

The simplest way to contribute changes to the Shellfish project is through github.
The process is described [here](https://guides.github.com/activities/contributing-to-open-source/).
If you're unfamilair with the process, contact me either throught he Issues tab or through
email and I'll walk you through what you need to do.

### Contribution guidelines

- Test your code thoroughly. (The standard Go testing files are great, but not neccessary.)
- Run the `go fmt` command on your code before contributing.
- All public functions, types, and variables/constants need to have a comment describing
what they do unless it is very obvious. There's no need to make them long or to use
any particular type of formatting. Put them on the line before the
function/type/variable so that automatic documentation-reading tools know what do with them.
- All packages need some sort of top-level comment ([like this](https://github.com/phil-mansfield/shellfish/blob/master/parse/config.go#L1)). Put it
on the line before the package name in one of the files in the package so automatic
documentation-reading tools know what to do with them.

## Language/Project Quirks

### Debugging Print Statements

Because Shellfish uses `stdout` to communicate between different processes, you will want
to make sure that all your print statements print to `stderr`. The simplest way to do this
is with the `log` package. At the top of a file, add `"log"` to the import list and use
`log.Printf(...)`

### Performance Profiling

Go supports gprof-like profiling. A long-winded (but good) description of how to use them
with example code can be found [here](https://blog.golang.org/profiling-go-programs).
