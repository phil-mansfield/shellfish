# Quickstart

Shellfish is a collection of programs which allow you to analyze splashback
shells. Half of them help you calculate splashback shells and half of them help
you manage and analyze merger trees. Any one of these programs can be used without
the other, so this quickstart will focus on  only on the splashback shell
calculation programs.

## Setup

### Make a Configuration File

The first step to using Shellfish is to create a configuration file and tell
Shellfish where it is. To get an example config file, type `$ shellfish help config`.
This example will contain a file with a number of variables and comments telling you
what to set them to. For now, set `HaloType = nil` and `TreeType = nil` and delete
any variables that start with the word `Halo` or `Tree` or which are commented out
to begin with. They should all be self explanitory with one exception,
`SnapshotFormat` and `SnapshotFormatMeanings`. These are a neccessary evil because
there are a lot of different ways that peoplelike to format their simulation output.

You can think of `SnapshotFormat` and `SnapshotFormatMeaning` as describing a call
to `sprintf` where the result is the name of one of your catalog files. For example,
imagine that your simulation was broken up into a hundred folders titled
`snapshot_<snapshot index>/` and the particles from that snapshot were broken up
into two hundred files called `particles_<block index>.dat`.  An example file in this
structure would be `path/to/sim/snapshot_11/particles_66.dat`. You would set the
relevant variables in your config file the following:
```
SnapshotFormat = path/to/sim/snapshot_%d/particles_%d.dat
SnapshotFormatMeaning = Snapshot, Block
SnapMin = 1 # These bounds are inclusive
SnapMax = 100
BlockMins = 1
BlockMaxes = 200
```
You can also handle more complicated file layouts. Say your files looked like
`path/to/sim/snapdir_013/snapshot_013.8.7.12.dat`. Here we have repeated variables,
zero-padding, and multiple block variables (maybe this file name is specifying a
location in Lagrangian space). This is what your configuration variables would look like
```
SnapshotFormat = path/to/sim/snapdir_%03d/particles_%03d.%d.%d.%ddat
SnapshotFormatMeaning = Snapshot, Snapshot Block0, Block1, Block2
SnapMin = 1 
SnapMax = 100
BlockMins = 0, 0, 10 # The range for each block variable can be different
BlockMaxes = 10, 10, 30
```

This is powerful enough to specify most directory structures, although there are some
reasonable ones which it will miss (for example, if you included the scale factor in
the the directory names in addition to the snapshot index). [Here](https://github.com/phil-mansfield/shellfish/blob/master/doc/snapshot_format_workaround.md) is a way to work around this.

(If you correctly specified `SnapshotFormat` and `SnapshotFormatMeaning`, congratulations:
that's the most complicated part of using Shellfish and you only need to do it once. It's
all downhill from here.)

### Tell Shellfish About Your Configuration File

Before running Shellfish, you need to tell it where your config file is. You do this
by setting the environment variable `SHELLFISH_GLOBAL_CONFIG`. 
```bash
export SHELLFISH_GLOBAL_CONFIG=path/to/my.config # If using bash
setenv SHELLFISH_GLOBAL_CONFIG path/to/my.config # If using csh
```
(If you don't know what unix shell you're using, run `echo $0`)

You will need to do this every time you open a new terminal and at the top of
all your shell scipts. If you are only ever going to run Shellfish on a single
simulation you may want to add this line to your `.bash_rc` or `.profile` file
for the duration of your research project.

## Analysis

Now we can move on to the fun part: using Shellfish to calculate splashback
shells. To follow along, go to one of your halo catalogs, pick out your
favorite cluster, and write down its snapshot index, X, Y, and Z coordinates,
its R200m value, and some sort of ID for that halo (it doesn't matter what it
is).

### Finding Shells

THe most important program in Shellfish is its shell finder (called, conveniently
enough, "shell"). This program can be run by calling `shellfish shell`. Try running
it. You'll notice that nothing happens. This is because shell, like most Shellfish
programs is expecting input from `stdin`. Specifically, it's expecting a
space-separated ASCII table where the first column is the halo ID, the second is the
snapshot index, the third through fifth are the X, Y, and Z coordinates in comoving
Mpc/h, and the sixth is R200m in the same units.

The way we pass information to `stdin` of a program is through the unix pipe, `|`.
Specifically, we call a program which will print out this type of table, write the
pipe symbol, and then call `shellfish shell`. The simplest such program is `echo`,
which just prints out its input. Here's what this looks like for one of the bigger
clusters in one of my simulations:
```bash
echo "80431577 100 13.6225 86.3578 53.1017 0.815028" | shellfish shell
```
After waiting about a minute (this halo had a million particles and I was only using
a single thread), I get output that looks like this:
```
# Column contents: ID(0) Snapshot(1) X [cMpc/h](2) Y [cMpc/h](3) Z [cMpc/h](4) R200m [cMpc/h](5) P_ijk(6-23)
80431577 100 13.6225 86.3578 53.1017 0.815028 0.979897 -0.0616729 0.35461 0.122959 0.281588 -0.0869048 -0.0560629 0.0831245 0.244373 -0.0405277 -0.0488017 -0.302187 -0.61362 0.357011 2.46993 0.0243595 0.525989 2.74402
```
The first line is a comment describing the contents of the output table and the second is
the data (if we had multiple lines in the input table, we would have had multiple lines in
the output table). The first six columns are your input data and the remaining columns
specify the shell shape in terms of [Penna-Dines coefficients](https://github.com/phil-mansfield/shellfish/blob/master/doc/penna_coefficients.md) (think of them as slightly
modified spherical harmonics... but don't worry: you won't need to use them). These
contain all the information about the shell shape, and in principle this is all you need
to do any analysis you want.

You don't just have to use `echo` to send input to shellfish programs. If you have a file
containing an input table, you can use `cat` to print it:
```bash
cat halo_coordinates.dat | shellfish shell
```
You can also write a script that extracts coordinates from some big complicated halo
catalog and print them out. A toy example might look something like this:
```python
# Python file called get_coordinates.py

ids = list_of_ids_I_want()
for id in ids:
    snap, x, y, z, r200m = parse_my_halo_catalog(id)
    print id, snap, x, y, z, r200m
```
which can then be used like this
```bash
python get_coordinates.py | shellfish shell 
```

Shellfish offers its own halo catalog parser, but it's not neccessary to use it.

### Calculating Shell Properties

Having the Penna-Dines coefficients is all well and good, but they're somewhat of
a pain to use. Shellfish offers another program, stats, which generates a catalog of
useful parameters. You can call it via `shellfish stats`. Much like the shell, stats
takes input from stdin. As luck would have it, it wants an input catalog that takes
exactly the same form as the output catalog of shell, meaning that we can just pipe
the two together.

I ran stats on that halo we looked at in the previous section like this:
```bash
echo "80431577 100 13.6225 86.3578 53.1017 0.815028" | shellfish shell | shellfish stats
```

The output looked like this:
```
# Column contents: ID(0) Snapshot(1) M_sp [M_sun/h](2) R_sp [cMpc/h](3) Volume [cMpc^3/h^3](4) Surface Area [cMpc^2/h^2](5) Major Axis [cMpc/h](6) Intermediate Axis [cMpc/h](7) Minor Axis [cMpc/h](8) Ax(9) Ay(10) Az(11)
80431577 100 4.1202e+13 1.13591 6.13931 18.3738 1.50995 1.11079 1.01605 -0.928776 -0.35794 -0.096201
```
Looking at columns 2 and three, we see the splashback mass of this halo is 4e13 M_sun/h
and the splashback radius is 1.1 comoving Mpc/h.

### Learning More

If you would like a detailed technical explanation of all the programs in Shellfish
and an exact specification of their input and output you can call `shellfish help`.
The rest of the toolset is explained qualitatively with examples [here](https://github.com/phil-mansfield/shellfish/blob/master/doc/toolset.md),
along with a description of how to change program parameters. Other advanced topics
(most importantly memory and time constriants.)
