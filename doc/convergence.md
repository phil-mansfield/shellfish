# Shellfish's convergence limits

There are two convergence limits which you should be aware of before using Shellfish
for paper-caliber scientific work: the particle convergence limit and the accretion
rate convergence limit.

The particle convergence limit is simple: below 50,000 particles, Shellfish will
systematically underestimate shell radii by at least 5% and this amount rapidly
increases as particle counts decreate

The accretion rate limits are more complicated. Shellfish encounters at least 10%
underestimation of shell sizes for _very_ low accretion rates. If we define accretion
rates to be Gamma = (ln(M(a1)) - ln(M(a2))) / (ln(a1) - ln(a2)) for z = 0, 0.5, 1, 2, and 4,
this cutoff occurs at Gamma = 0.5. This cut will remove about 20% of Milky Way mass halos
at z = 0 and a negligible amount at all larger mass ranges and at all other redshifts.
(This is because these halos are actually _losing_ mass depsite their positive accretion
rate: the argument can be found in our paper.)

There is also an upper mass limit that should be used when doing statistical studies,
although this has nothing to do with Shellfish and everything to do with finite box sizes.
If you want to look at highly accreting halos, you need to first ensure that the halo
accretion rate distribution at the maximum mass scale that you look in a given box is
consistent with the accretion rate distribution at that mass scale in larger boxes.
