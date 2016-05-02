# Setting up `SnapshotConfig`

It is an unfortunate fact that the directory strucutre of different simulations can
be vastly different. Because Shellfish automatically traverses your directory strucutre,
it's a neccessary evil that some time needs to be spent specifying the strucutre in
the config file.

In almost all cases, simulation output can be described by a series of nested loops.
Suppose, for example I have a one hundred snapshot ARTIO simulation where every
snapshot exists in its own directory: `snapshot_99`, `snapshot_0`, and the like.
And inside of each directory are 50 data files data files (plus a header) labeled
`my_simulation_snap99_a0.1041.p001`, `my_simulation_snap0_a1.0001.p050` etc. The
following C for loop would
print out the name of each of my output files:

    double scaleFactors[] =  { ... };
    for (int snapshot = 0; snapshot < 100; snapshot++) {
        for (int block = 1; block <= 50; block++) {
            printf("snapshot_%d/my_simulation_snap%d_a%.4f.p%03d",
                   snapshot, scaleFactors[snapshot], snapshot, block);
        }
    }

When specifying your snapshot names in the Shellfish config file, you are specifying
this for loop. You pass the config file a format string and you tell it what each of
those loops are.

To specify the files that I described above, first you set the format string:

    SnapshotFormat = "snapshot_%d/my_simulation_snap%d_a%s.p%03d"
    
(Note the change from `%.4g` to `%s`. More on that later.) Next, you tell Shellfish
what each of those formatting verbs mean:

    SnapshotFormatMeanings = Snapshot, Snapshot, ScaleFactor, Block

Next, you specify the ranges taken on by each of these values:

    SnapMin = 0
    SnapMax = 99
    BlockMins = 1
    BlockMaxes = 50
    ScaleFactorFile = my_list_of_scale_factors.txt

Bounds are inclusive. `ScaleFactorFile` is a plaintext file containing the scale
factors of each snapshot in the range you want to analyze. You should extract
these values directly from the output filenames themselves. (I've provided an [example Python script](https://github.com/phil-mansfield/shellfish/blob/master/doc/scale_factor_ex.py)
to show how to do this.) This file must have the same number of lines as `SnapMax - SnapMin + 1`.

Even if you do your snapshot differentiating entirely through scale factors and don't use
explicit snapshot indices anywhere in your file names, you still need to set `SnapMin` and
`SnapMax` accordingly.

# Special Format-Specific Considerations

### ARTIO

If using ARTIO snapshot formats, `SnapshotFormat` should specify
only the names of the data files themselves, not the name of the header.

Shellfish supports adaptively refined masses for ARTIO formats.

### LGadget-2

Shellfish does not currently support adaptively refined particle masses
for LGadget-2 formats. (However, it will be straightforward to add,
if requested.)

### gotetra

(Coming soon!)
