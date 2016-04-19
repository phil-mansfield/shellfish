# Walkthrough

### Anatomy of Shellfish Commands

Shellfish is set up as a number of independent command-line tools which read
input from stdin, write it to stdout, and can be piped together as needed. Each
tool is a different command line mode, run as

     $ shellfish [mode-name] [...]

Each one of these modes takes 1-2 configuration files as arugments. Every mode
takes a "global" configuration file as a first argument, and some of the more
complicated modes can take a second, optional, mode-specific configuration file.
The global config file describes the simulation being analyzed, and the optional
config files give mode-specific details.

For example, the shell mode (the most important mode in the tool chain) can
be run like this:

    $ shellfish shell example.config

Or, if an experienced user wants finer control, they could run it like this:

    $ shellfish shell example.config my.shell.config

We will cover the contents of these configuration files later in this 
walkthrough.

### Making Things More Concise

If you know in advance that all your Shellfish analysis will be done on the
same simulation for the forseable future (which will usually be the case), you
can ellide the global config file from every call by setting the environment
variable `$SHELLFISH_GLOBAL_CONFIG`. If done, the two calls from above
become the more concise:

    $ shellfish shell
    $ shellfish shell my.shell.config

(For those not familiar with unix programming, the way this is done depends on
what shell you are running. You can find out what shell you are running by
typing `$ echo $0`. If you are running `sh` or `bash`, you can set this 
variable by executing `$ export = SHELLFISH_GLOBAL_CONFIG=example.config`.
If you are running `csh` or `tcsh`, you can set this variable by executing
`$ setenv SHELLFISH_GLOBAL_CONFIG example.config`. You can check what the
the variable is currently set to by typing `$ echo $SHELLFISH_GLOBAL_CONFIG`.)

As long as this variable is set, global config files _cannot_ be passed from the
command line. If you would like to be able to pass them from the command line
again, set `$SHELLFISH_GLOBAL_CONFIG` to an empty string.

### Example Calls

For most users, the series of Shellfish calls they will have to make will look
almost exactly the same.

Here is an example Shellfish invocation by a user who has already set
`$SHELLFISH_GLOBAL_CONFIG` and wants to find the splashback radii for a family
of Milky Way-sized halos:

    $ shellfish id mw_halos.id.config |
        shellfish coord |
        shellfish shell |
        shellfish stats > my_output.dat

(For those not familiar with unix shell programming, each `|` is a unix pipe
which feeds the stdout of earlier processes into the stdin of later processes,
and the `>` writes the last process's output to a file named `my_output.dat`.
I put all of them on different lines here, but you can write all of this as a
one-line command.)

Let's break this down. First, `shellfish id` generates a family of IDs
depending on the specifications supplied in `mw_halos.id.config`. This is
passed to `shellfish coord`, which finds the location of the halos specified
by these IDs. This is passed to `shellfish shell`, which identifies shells for
each of these halos. The Penna coefficients corresponding to these shells are
passed to `shellfish stats`, which uses them to calculate splashback radii, masses, and minimum/maximum radii.

I ran this command on the `L0063` simulation suite described in
Diemer & Kravtsov 2014 and wrote `mw_halos.id.config` in a way that requested
the 1005th, 1006th, and 1009th most massive halos in the z=0 snapshot of the
simulation (I'll explain how to do that later). I got output that looked like
this:

    # Column contents: ID(0) Snapshot(1) M_sp(2) R_sp(3) R_sp,min(4) R_sp,max(5)
    169665239 100 1.373e+12 0.4084 0.3121 0.5297
    168208646 100 1.244e+12 0.3683 0.3129  0.429
    168863226 100 1.284e+12 0.3576 0.2613 0.4081

The ouput consists of a description of each column followed by three rows which
each correspond to a different halo. The first two columns are identifying 
information: an ID taken from the halo catalog and the index of the snapshot
that the halo is in (The `L0063` simulation has 101 snapshots, so z=0
corresponds to an index of 100). The next four columns give the mass contained
in the splashback shell, the volume-weighted splashback radius of the shell, and
the minimum and maximum radii that the shell reaches.

Commands can be added or removed from the chain to modify the calculation and
the output. For example, this set of commands will track the halos back through
previous snapshots in the simulation and will output the evolution of the
splashback shells over time.

    $ shellfish id mw_halos.id.config |
        shellfish tree |
        shellfish coord |
        shellfish shell |
        shellfish stats > my_output.dat

This set of commands will output the parameters of the Penna function, which
would allow you to perform your own analysis on the shell:

    $ shellfish id mw_halos.id.config |
        shellfish coord |
        shellfish shell > my_output.dat

This is what typical call looks like, but to run this, we still need to 

### Config File Input

Every Shellfish config file has the same basic form:

    [my_mode.config]
    # Comment
    var1 = value1 # Inline comment
    var2 = list, of, values

At a minimum, every user must write a "global" config file that tells Shellfish
where to find snapshots, halo catalogs, etc.

# Advanced Options

### Selecting Halos Without `shellfish id`

### Using Unsupported Halo Catalogs and/or Merger Trees

### Command Line Options

Coming soon!

### Full Mode Specifcations

Coming soon!