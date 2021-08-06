# Shellfish Resource Consumption

### Caching
The first time Shellfish is run on a particular snapshot of a simulation, it
will do some full-box analysis needed to accelerate future analysis commands.
This includes collecting the headers of every block in a single place and
and parsing halo catalogs 
This can take some time depending on the underlying file format and simulation
size. This will take a minute or three per snapshot for a billion particle
simulation and will require a few GB of RAM. It will generate files which are
a small fraction of the size of your halo catalogs.

This setup step will be performed every time you use a new global configuration file.

### Memory

With default settings and an O(100)-sized halo catalog, Shellfish consumes about
57 MB of RAM per halo with default parameters. The vast majority of this comes from
maintaining ~100,000 line of sight profiles per halo, which can't be reduced without
changing parameters in `shell.config`, although some of this comes from heap
fragmentation and the overhead of loading in snpashots and causes the memory overhead to
very slowly increase of time. This increase is largely unimportant unless you are trying
to stay _exactly_ under your node's memory limit.
If you encounter cases where
significantly more than 57 MB are being used per halo for a large number of halos,
[submit a bug report](https://github.com/phil-mansfield/shellfish/issues).

In later versions of Shellfish, you will be able to specify a maximum memory limit for Shellfish,
and it will do all its analysis without exceeding this limit, regardless of halo count.
For the time being, if you need to analyze a large number of haloes, you will need to manually
split up the input catalogs.

### CPU Time

Rigorous benchmarks are coming soon.

(On my machine with the default `shell.config` parameters, in one hour Shellfish can analyze
a colleciton of halos which together contain 7 million particles within their R200m
spheres, regarless of their size. Parallelization is good, meaning that you can speed this
up by a factor of 16 by running Shellfish on a node with 16 cores available.
If you find that you are getting significantly worse performance
on your machine, [let me know](https://github.com/phil-mansfield/shellfish/issues).)

Shellfish does not currently support MPI. Are parallelism is thread-based and done
on a single node. In fact, due to the way that caching works,
you can't even run two Shellfish processes using the same configuration file on
two different nodes simulataneously (unless you're sure caching is finished, which you
might be.)