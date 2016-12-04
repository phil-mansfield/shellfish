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
as well as heap fragmentation. In practice you are safe assuming
that when analyzing hundreds or thousands of halos, the overhead will not exceed twice
the minimum (almost always much less than this). If you encounter cases where more than
25 MB are being used per halo for a large number of halo,
[submit a bug report](https://github.com/phil-mansfield/shellfish/issues).

This ~13-25 MB per halo limit is per-snapshot, meaning that tracking a thousand halos
across a hundred snapshots will consume 13-25 GB, not multiple TB.

In version 1.1.0, you will be able to specify a maximum memory limit for Shellfish,
and it will do all its analysis without exceeding this limit, regardless of halo count.
If you need to analyze a large number of halos now, you will need to manually split up
the input catalogs.

### CPU Time

Rigorous benchmarks coming soon!

(On my machine with the default `shell.config` parameters, single-threaded analysis
takes about eight seconds for every 100,000 particles within the R200m radii of all
the analyzed halos. If you find that you are getting significantly worse performance
on your machine, [let me know](https://github.com/phil-mansfield/shellfish/issues).
This means that Shellfish runs at about the same speed at Rockstar for the same
number of particles.)

**Note**: This is only true for isolated halos. Small halos which are close to clusters
can take as long to analyze as the clusters themselves. The real figure of merit is how
many particles are contained within 3*R200m, a value which is only loosely correlated
with how many particles are contained within R200m. However, most halo catalogs do not
contain this information.
