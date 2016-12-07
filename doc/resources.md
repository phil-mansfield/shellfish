# Shellfish Resource Consumption

### Caching
The first time Shellfish is run on a particular snapshot of a simulation, it
will do some full-box analysis needed to accelerate future analysis commands.
This includes collecting the headers of every block in a single place and
and parsing halo catalogs 
This can take some time depending on the underlying file format and simulation
size. This will take a minute or three per snapshot for a billion particle
simulation and will require a few GB.

This setup step will be performed every time you use a new global configuration file.

### Memory

With default setting, Shellfish consumes about 13 MB per halo along with some
hard-to-model overhead due to storing portions of the underlying particle snapshots
as well as heap fragmentation. In practice, you are safe assuming
that when analyzing hundreds or thousands of halos, the overhead will not exceed twice
the minimum (almost always much less than this). If you encounter cases where more than
25 MB are being used per halo for a large number of halo,
[submit a bug report](https://github.com/phil-mansfield/shellfish/issues).

More predictable memory usage is the number one issue for version 1.1.0 of Shellfish.

Also in version 1.1.0, you will be able to specify a maximum memory limit for Shellfish,
and it will do all its analysis without exceeding this limit, regardless of halo count.
If you need to analyze a large number of halos now, you will need to manually split up
the input catalogs.

### CPU Time

Rigorous benchmarks coming soon!

(On my machine with the default `shell.config` parameters, in one hour Shellfish can
a colleciton of halos which together contain 7 million particles within their R200m
spheres, regarless of their size. Parallelization is good, meaning that you can speed this
up by a factor of 16 by running Shellfish on a node with 16 cores available.
If you find that you are getting significantly worse performance
on your machine, [let me know](https://github.com/phil-mansfield/shellfish/issues).
This means that Shellfish runs at about the same speed at Rockstar for the same
number of particles.)

Shellfish does not currently support MPI. Are parallelism is thread-based and done
on a single node. In fact, due to the way that caching works,
you can't even run two Shellfish processes using the same configuration file on
two different nodes simulataneously (unless you're sure caching is finished, which you
might be.) There are plans to [add MPI support](https://github.com/phil-mansfield/shellfish/issues/128)
by version 1.1.0.
