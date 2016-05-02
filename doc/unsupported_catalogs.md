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

	# Columns: ID(0), Snapshot(1), X [Mpc/h](2),
	# Y [Mpc/h](3), Z [Mpc/h](4), R200m[Mpc/h](5)
	5234987 100 42.742 189.001   5.241 1.023
	 100772  55  0.111 100.511 150.226 0.500
	6709823 100 15.091   7.123  88.178 0.441
	
The output halos will be in the same order as the halos in your original text file.

*Caveat*: Make sure you don't pass Shellfish the location of subhalos and if your
simulation has periodic boundary conditions, make sure each coordinate has been
transformed back into the [0, L) range.
