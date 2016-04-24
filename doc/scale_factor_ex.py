import glob

# This file shows an example of how to extract the scale factors in the file
# names of your simulation in a few lines of python. More complicated naming
# conventions will require slightly more complicated extraction scripts.
#
# Specifically, it will read the example files I put in this directoy which are
# of the form:
#
#    example/path/to/files/snapshot_a<scale-factor>/block_<block-index>.dat

# Unix-like filename pattern:
pattern = "example/path/to/files/snapshot_a*/block_0.dat"
# Identifying strings both before and after the wildcard (you could also do this
# by just hard coding the index offsets):
string_before, string_after = "snapshot_a", "/block_"

fnames = glob.glob(pattern)

for fname in fnames:
    start = fname.find(string_before) + len(string_before)
    end = fname.find(string_after)
    print fname[start: end]
