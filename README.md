# Shellfish

Shellfish (**SHELL** **F**inding **I**n **S**pheroidal **H**alos) is a toolchain for finding
the splashback shells of individual halos within cosmological simulations. Shellfish
is written in the programming langue [Go](https://golang.org/).

Shellfish is currently in version 0.2.0 with the next planned release being version
0.3.0. Its conifguration file API is mostly stable, it is not yet fully documented,
and does not yet have a C-compatible API. The current maintainer in Phil Mansfield
at the University of Chicago.

### Installation

There are two steps to installing Shellfish. The first is installing a Go compiler
(which is relatively painless compared to installing most other compilers), and the
second is compiling Shellfish and its dependencies (which is also painless).

If you run into significant problems during installation, you can ask me for help.
I go over the best way to do this in the next section.

If you are working on a computer that you own, download the latest version of the Go
compiler from [here](https://golang.org/doc/install) and following the instructions.
If you are working on a cluster, you can also ask the cluster staff to install the
compiler for you. Regardless of how you install it, you will need to make a few changes
to your `.profile` file, so make sure to read the section titled *Test your installation*
and run the hello world program there.

Once you've done this, its time to download and install Shellfish's dependencies. Type
the following commands into your console:

	$ go get github.com/gonum/matrix
	$ go get github.com/phil-mansfield/consistent_trees
	$ go get github.com/phil-mansfield/go-artio
	$ go get github.com/phil-mansfield/shellfish
	
Lastly, type

	$ go install github.com/phil-mansfield/shellfish

And you're done!

### Getting Help

The best way to get help is to submit an issue on this project's
[Issues page](https://github.com/phil-mansfield/shellfish/issues). This
is a page that collects all the bug reports, feature requests, and
help requests in the same place. You will need a github account, but
[signing up](https://github.com/join) is quick.

You can submit an issue by clicking the green button titled "New Issue"
in the upper right corner of the issues page. This will open a submission
form where you can describe the problem you are encountering. Feel free to
select descriptive tags from the panels on the right of the form, but if
you don't, I will handle it for you.

The standard suggestions for reporting software bugs/problems applies:
* Check beforehand to see if someone else already encountered your problem.
(Look [here](https://github.com/phil-mansfield/shellfish/issues?q=is%3Aopen+is%3Aissue+label%3Abug)
for bugs, [here](https://github.com/phil-mansfield/shellfish/issues?q=is%3Aopen+is%3Aissue+label%3Aenhancement)
for feature requests, [here](https://github.com/phil-mansfield/shellfish/issues?utf8=%E2%9C%93&q=is%3Aissue+label%3A%22help+wanted%22) for generic help requests and [here](https://github.com/phil-mansfield/shellfish/issues?utf8=%E2%9C%93&q=is%3Aissue+label%3Aquestion+) for questions)
* Make sure you are using the most up to date version of Shellfish (check your
version by typing the command `$ shellfish version`).
* Provide all information relevant to your problem (shell commands, config files, manual
input catalogs if relevant, etc.)
* If you encounter your problem when running a complex set of operations, try to find the
simplest possible configuration which still exhibits the problem.

### How to Use Shellfish

Because of the nature of the problem it is trying to solve, Shellfish works differently
from most other halo catalog tools. In particular, it is designed as a family of isolated
tools which communicate with one another through I/O redirection, in the same way that
unix utilities often do. This allows the average user to run Shellfish without excessive
configuration while also allowing advanced users to do more complicated tasks.

You can find a tutorial on using Shellfish [here](https://github.com/phil-mansfield/shellfish/blob/master/doc/tutorial.md).
It typically takes about 10 minutes to read through.

Future versions of Shellfish (0.3.0 and up) will be packaged with a
`simple_shellfish` tool which will perform the most common sequences of commands
automatically.

### List of Supported File Formats

Currently supported particle catalog types:

* gotetra
* LGadget-2
* ARTIO

Currently supported halo catalog types:

* All text column-based catalogs

Currently supported merger tree types:

* consistent-trees

If you would like to use a particle catalog type which is not supported here,
plase submit an Issue requesting support. Shellfish is written in a way that
makes it easy to interact with unsupported halo catalogs and meger trees (as
covered in [the tutorial](https://github.com/phil-mansfield/shellfish/blob/master/doc/tutorial.md)),
but feel free to submit an issue requesting support as well. When submitting
an issue asking for support of other file formats, please make sure to link
to either a specification of the file format or to a C or Go library
which reads from this format.

### Useful Background Reading

Coming soon!

### Contributing to Shellfish

Coming soon!
