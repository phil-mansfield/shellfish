# Using the Full Shellfish Toolset (By Example)

This document assumes that you've already read the [quickstart](https://github.com/phil-mansfield/shellfish/blob/master/doc/quickstart.md)
tutorial where I described the simplest way to use Shellfish. Here, I'll talk
about the other useful features that Shellfish offers. This includes using
Shellfish's halo catalog reader, configuring shellfish tools, saving partial output, and calculating various
types of splashback-related halos.

## Using Shellfish's Catalog Reader

In the quickstart tutorial we saw how to use Shellfish once you've prepared a halo
catalog that has IDs, snapshot indices, halo positions, and halo radii as the first
six columns:

```bash
cat my_halo_catalog.txt | shellfish shell | shellfish stats > my_splashback_catalog.txt
```

While this works, it can be a bit annoying for five reasons: first, whatever halo catalogs
you are using probably don't come in this column order naturally; second, Shellfish's
[memory constraints](https://github.com/phil-mansfield/shellfish/blob/master/doc/resources.md)
as of version 1.0.0 will force you to run it on no more than a few thousand halos at a time
(an issue which is [currently being worked on for version 1.1.0](https://github.com/phil-mansfield/shellfish/issues/127));
third, Shellfish's [convergence limits](https://github.com/phil-mansfield/shellfish/blob/master/doc/convergence.md);
fourth, if you want to look at the accretion history of a halo, you'll need to read
the main progenitor line from a halo merger tree, which can be an involved task; and
fifth, parsing halo catalogs is slow: doing it effectively requires caching which
can be annoying to write yourself.

Because of this, Shellfish comes bundled with the catalog/tree reader that I
used while developing the shell finder. It won't work on all catalogs, but will work on most
of them.

To use the catalog reader, you will need to update the file pointed to by
`$SHELLFISH_GLOBAL_CONFIG` so that all the variables prefixed by `Halo` are
filled out. You can get an example config file with comments explaining every
variables by typing the command `shellfish help config`.

Once this is done, the simplest usage of the catalog reader looks like this:
```bash
cat my_halo_ids.txt | shellfish coord | shellfish shell | shellfish stats > my_splashback_catalog.txt
```
where `my_halo_ids.txt` is a catalog containing halo IDs at the first column
and snapshot indices as the second column. The program coord will generate a catalog of
halo positions that is exactly the format that the shell program needs.

If you don't want to generate an ID catalog yourself, Shellfish can do it for you. The
most useful way to do this is the following:
```bash
shellfish id --M200mMin 1e14 --M200mMax 2.5e15 --Snap 100 |
    shellfish coord |
    shellfish shell |
    shellfish stats > my_splashback_catalog.txt
```
That first line will parse the 100th snapshot in your simulation, and extract
all the halos with 1e14 M_sun/h < M200m < 2.5e15 M_sun/h. It will also remove halos
which are highly likely to be within the splashback shell of a larger halo. (Since
shells haven't been calculated yet, this is done heuristically. Support for identifying
subhalos correctly is an open issue for Shellfish 
version 1.1.0). The remaining lines pass those IDs through the rest of the pipeline.

If you want to track every halo back in time, fill out the variables that start with 
`Tree` in the file `$SHELLFISH_GLOBAL_CONFIG` points to.
```bash
cat my_halo_ids.txt | shellfish tree | shellfish coord |
    shellfish shell | shellfish stats > my_splashback_catalog.txt
```
(the first command can, of course, be replaced by a call to `shellfish id`).

## Configuring Shellfish Tools

In addition to the global configuration file, every Shellfish program takes
an optional, program-specific configuration file. These config files fill the same role
as command line flags. For example, the command 

```bash
shellfish id --IDStart 500 --IDEnd 1000 --Snap 100 |
    shellfish coord |
    shellfish shell |
    shellfish stats > my_splashback_catalog.txt
```
(which selects the 500th - 1000th largest halos)

could be written as

```bash
shellfish id 500_1000.id.config |
    shellfish coord |
    shellfish shell |
    shellfish stats > my_splashback_catalog.txt
```

where `500_1000.id.config` is the file
```
[id.config]
IDType = M200m
IDStart = 500
IDEnd = 1000
Snap = 100
```

You can find the full list of configuration variables for each program through
Shellfish's help mode (e.g. `shellfish help id.config`). If you supply both a
configuration file and command line arguments and they set a variable to two
different values, the command line flags will win.

## Other Shellfish examples

### Keeping subhalos in the output of ID

```
shellfish id --M200mMin 1e14 --M200mMax 2.5e15 --ExclusionStrategy none
```

### Extracting arbitraty coordinates with coord

If you want coord to know about variables other than the ones needed by shell
(for example, if you want to extract catalogs for other analysis), you can update
config file variables to tell it about other variables. For example, this is what
the relevant part of my configuation file looks like

```
HaloValueNames = Scale, ID, M200m, R200m, Rs, VMax, X, Y, Z, Vx, Vy, Vz, MVir, M200c, M500c, M2500c, BToA, CToA, Ax, Ay, Az, MPeak, VPeak
HaloValueColumns = 0, 1, 10, 11, 12, 16, 17, 18, 19, 20, 21, 22, 39, 40, 41, 42, 46, 47, 48, 49, 50, 61, 62 
HaloValueComments = "", "", Msun/h, ckpc/h, ckpc/h, pkm/s, cMpc/h, cMpc/h, cMpc/h, pkm/s, pkm/s, pkm/s, Msun/h, Msun/h, Msun/h, Msun/h, "", "", ckpc/h, ckpc/h, ckpc/h, Msun/h, pkm/s
```

(Aside from `ID`, `X`, `Y`, `Z`, and `M200m`, the names you use are not important.)

You can then generate catalogs in the following way:
```bash
cat my_halo_ids.txt | shellfish coord --Values "X, Y, Z, Vx, Vy, Vz"
```

### Only extracting a subset of snapshots with tree

Suppose you wanted to track a group of halos back in time, but only needed a subset of all the
snapshots that the halos exist within. This can be done as follows:
```bash
cat my_halo_ids.txt | shellfish tree --SelectSnaps "36, 47, 64, 77, 87, 100"
```

### Constructing median profiles

The program prof can be used to compute a variety of halo profiles. The most useful
for splashback analysis is the so-called angular median density profile. An example usage
is the following:

```bash
cat my_halo_ids.txt | shellfish coord | shellfish prof
```
