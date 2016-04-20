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
become more concise:

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
passed to `shellfish stats`, which uses them to calculate splashback radii, 
masses, and minimum/maximum radii.

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

### Config File Input

Every Shellfish config file has the same basic form:

    [my_mode.config]
    # Comment
    var1 = value1 # Inline comment
    var2 = list, of, values

At a minimum, every user must write a "global" config file that tells Shellfish
where to find snapshots, halo catalogs, etc. You can get a skeleton config file
complete with comments explaining every variable by running

	$ shellfish help config

Go through that skeleton file line-by-line and set each variable to the values
that correspond to your simulation. Below I've copied the global config file
that I used in the previous example if you want to see what a working file looks
like:

	[config]
	
	# File formats
	
	SnapshotType = gotetra
	HaloType = Rockstar
	TreeType = consistent-trees
	
	# Directories
	
	HaloDir = /project/surph/diemer/Box_L0063_N1024_CBol/Rockstar200m/hlists
	TreeDir = /project/surph/diemer/Box_L0063_N1024_CBol/Rockstar200m/trees
	MemoDir = /project/surph/mansfield/data/sheet_segments/Box_L0063_N1024_G0008_CBol/gtet_memo
	
	# Snapshots
	
	SnapshotFormat = /project/surph/mansfield/data/sheet_segments/Box_L0063_N1024_G0008_CBol/snapdir_%03d/sheet%d%d%d.dat
	SnapMin = 6
	SnapMax = 100
	FormatMins  = 0, 0, 0
	FormatMaxes = 7, 7, 7

The only confusing variable is `SnapshotFormat`, which is used to specify the
names of your particle snapshots. This is a neccessary evil that comes from
the wide range of file naming conventions that different simulations use. The
idea is to write a format string (like the one used by `printf`) which will take
a snapshot specifier and some arbitrary number of block identifcation indices
as arguments, with the specifics depending on the exact simulation.

The other config file that most users will need to set is the
`shellfish id`-specifc config file. Its sole purpose to concisely communicate to
Shellfish what halos you want IDs from. A skeleton config file can be generated
by typing the command `$ shellfish help id.config`. Here is the configuration
file that I used in the example above:

	[id.config]
	
	Snap = 100
	IDs = 1005, 1006, 1009
	ExclusionStrategy = overlap

Skeleton config files can be found for the other modes (if needed) by typing
the help command followed by the file extension of the config file you want,
i.e.

	$ shellfish help [ config | id.config | shell.config | stats.config ]

# Advanced Options

### Selecting Halos Without `shellfish id`

Although the config file system is convenient for contiguous mass ranges and
for small sets of IDs, in more complicated situations where one might want
more control over the IDs supplied. Here are a few examples of why you might
want to do this:

* You want to run Shellfish on a subsample of the halos in a given mass range.
* You want to look at the evolution of splashback shells, but some snapshots
have corrupted catalogs or are otherwise missing.
* You are writing your own user-level load balancer for Shellfish.

This is easy to do because Shellfish takes input in the form of text. Simply
create a text file which contains one line for every halo you want to analyze.
The first column of each line should be the ID of halo and the second column
should be the index of the snapshot that halo is in. Don't worry about spaces
or empty lines or alignment or character counts: Shellfish will
handle it cleanly. Next, pipe that text file into Shellfish in the place where
the `shellfish id` call would have gone.

So this call

	$ shellfish id example.id.config | shellfish coord | shellfish shell | shellfish stats
	
could be replaced with this

	$ cat my_id_file.txt | shellfish coord | shellfish shell | shellfish stats
	
where the contents of `my_id_file.txt` look something like this:

	# Columns: ID(0), Snapshot(1)
	5234987 100
	 100772  55
	6709823 100
	
The output halos will be in the same order as the halos in your original text file.

*Caveat*: Don't pass Shellfish the IDs of subhalos. It will analyze them without
crashing, but subhalos almost never have meaningful splashback shells, so you're
unlikely to get useful results.

### Using Unsupported Halo Catalogs and/or Merger Trees

Although Shellfish provides native support for certain types of halo catalogs and
merger trees through the `shellfish id` and `shellfish tree` modes, it is possible
to make it work on simulations which use unsupported halo catalogs with a small
amount of work on the user end. The `shellfish shell` mode takes plaintext as input
and makes no reference the underlying halo catalogs, so if you manually pass it the
locations of halos you want to analyze everythign will work correctly.

Specifically, create a text file which contains one line for each halo you
want to analyze. Each line should have six columns. The first should be an
identifying ID (Shellfish won't use it for anything, but it will help you
cross-reference the output catalog), the second should be the index of the snapshot
that the halo is in, the next three columns should be the x, y, and z coordinates
in units of Mpc/h, and the last column should be R200m in units of Mpc/h.
If you already have software which reads your halo catalogs, this type of file
should be fairly painless to create. Don't worry about spaces or empty lines or
alignment or character counts: Shellfish will handle it cleanly.

Lastly, pipe that file into the `shellfish` toolchain in the same place where
the `shellfish coord` comand would have gone.

So this call

	$ shellfish id example.id.config | shellfish coord | shellfish shell | shellfish stats
	
could be replaced with this

	$ cat my_coord_file.txt | shellfish shell | shellfish stats
	
where the contents of `my_coord_file.txt` look something like this:

	# Columns: ID(0), Snapshot(1), X(2), Y(3), Z(4), R200m(5)
	5234987 100 42.742 189.001   5.241 1.023
	 100772  55  0.111 100.511 150.226 0.500
	6709823 100 15.091   7.123  88.178 0.441
	
The output halos will be in the same order as the halos in your original text file.

*Caveat*: Make sure you don't pass Shellfish the location of subhalos and if your
simulation has periodic boundary conditions, make sure each corrdinate has been
transformed back into the [0, L) range.

### Command Line Options

Coming soon!

### Full Mode Specifcations

Coming soon!

### Multi-node paralellism

Coming soon!
