# Selecting Halos Without `shellfish id`

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
or empty lines or alignment or character counts or comments: Shellfish will
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

*Caveat*: Don't pass Shellfish the IDs of subhalos. In the best cases it will crash
and in the worst cases it will complete analysis seamlessly (which is bad because
subhalos never have meaningful splashabck radii.)
