# Working Around Problems With `SnapshotFormat`

The `SnapshotFormat` variable assumes that your snapshots and block files can be
indexed by some number of integers. While this is usually true, it doesn't have to
be (the most common example is labeling your snapshot directories by redshift).

Older prerelease versions of Shellfish used had config file variables which allowed
redshifts to be used in file names, but this didn't really solve the issue and was
causing internal issues, so it has been deprecated. The method I currently suggest is
just to make symbolic links (i.e. file shortcuts that essentially take up no
disk space) and to have the name of the links be of a type that Shellfish can dela with.

You can create a symbolic link using the unix command `ln -s file_name link_name` or
the Python function `os.symlink`. The
one subtlety is that both arguments should be _absolute_ paths instead of _relative_ paths.

Suppose I have my simulations in a home directory called `/path/to/home` that are called
`snapshot_1.00000`, `snapshot_z0.92781/`, etc. and inside I have a collection of block
files inside called `block.A`, `block.B`, etc. This wouldn't be able to be specified by
`SnapshotFormat`

Here is a python file which could generate the corresponding symbolic links. Feel free
to modify it to work for you. It assumes
the existence of two functions which you can write yourself `get_snapshot_dirs(home)`,
which finds the (absolute!) names of all the snapshot directories in `home` (and excludes whatever else
you have in that folder) and returns them as a list which is sorted from earliest in time
to latest in time, and `get_block_files(dir)`, which does
the same for block files, but has no requirements on order.
```python
import os
import os.path as path

home = "/path/to/home" # An absolute path

# Create the directory that these new files will go in.
link_home = path.join(home, "shellfish_links")
os.mkdir(link_home)

snapshot_dirs = get_snapshot_dirs(home)
for i, snapshot_dir in enumerate(snapshot_dirs):
    block_files = get_block_files(snapshot_dir)

    # Create a directory to hold the links in this snapshot
    # which has an easily indexed name.
    link_dir = path.join(link_home, "snapshot_%d" % i)
    os.mkdir(link_dir)
    
    # Create links for each block using an easily indexed name.
    for j, block_file in enumerate(block_files):
        link_name = path.join(link_dir, "block_%d" % j)
        os.symlink(block_file, link_name)
```

Using this file, you would set up `SnapshotFormat` with the variables
```
SnapshotFormat = /path/to/home/shellfish_links/snapshot_%d/block_%d
SnapshotFormatMeanings = Snapshot, Block
```
