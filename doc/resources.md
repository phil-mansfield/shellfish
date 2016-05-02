# Shellfish Resource Consumption

##### Setup
The first time Shellfish is run on a particular snapshot of a simulation, it
will do some full-box analysis needed to accelerate future analysis commands.
This can take some time depending on the underlying file format and simulation
size. This can take around ten-twenty minutes for a billion particle simulation
and will require a few GB. If you find that more than 8 GB are being consumed
during this step, you should [let me know](https://github.com/phil-mansfield/shellfish/issues).

This setup step will be performed very time the global configuration file is
changed. If your use case involves frequently changing this file and you are finding
the setup time to be inconveniencing, [let me know](https://github.com/phil-mansfield/shellfish/issues).

##### Memory

With default setting, Shellfish consumes about 13 MB per halo along with some
hard-to-model overhead due to storing portions of the underlying particle snapshots
as well as heap fragmentation. In practice you are safe assuming
that when analyzing hundreds or thousands of halos, the overhead will not exceed twice
the minimum (almost always much less than this). If you encounter cases where more than
26 MB are being used per halo for a large number of halo, [submit a bug report](https://github.com/phil-mansfield/shellfish/issues).

This ~13 MB per halo limit is per-snapshot, meaning that tracking a thousand halos across
a hundred snapshots will consume 13 GB, not 1.3 TB.

Future versions of Shellfish (0.3.0+) will naturally handle this internally, but for
now if you want to analyze more than thousands of halos simultaneously you may need to
split the analysis up into multiple calls manually.

##### CPU Time

Rigorous benchmarks coming soon!

(On my machine with the default `shell.config` parameters, analysis takes about
one or two CPU seconds for every 100,000 particles within the R200m radii of all
the analyzed halos. If you find that you are getting significantly worse performance
on your machine, [let me know](https://github.com/phil-mansfield/shellfish/issues).)

##### Parallelism

Shellfish will automatically detect the number of cores available to it on a single
node and will load balance to the best of its ability across all of them. If you
would like Shellfish to use a smaller number of cores, set the `Threads` configuration
variable to the number of cores you would like to use.

Shellfish does not internally support multi-node parallelism, but . As long as mass
is roughly equipartitioned across nodes, the load balancing will be reasonably optimal.

(An example of such a Python script is coming soon!)
