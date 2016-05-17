import scipy.interpolate as intr
import matplotlib.pyplot as plt
import numpy as np
import data

# Judge not lest ye be judged

c_rat = data.c_grid.flatten()
ac_rat = data.ac_grid.flatten()
bc_rat = data.bc_grid.flatten()

bc, ac = np.meshgrid(data.acs, data.bcs)
bc, ac = bc.flatten(), ac.flatten()
okay = bc > ac

ac, bc = ac[okay], bc[okay]
c_rat, ac_rat, bc_rat = c_rat[okay], ac_rat[okay], bc_rat[okay]

pts = zip(ac, bc)
n = len(data.acs)
g_bc, g_ac = np.mgrid[n/5:n, n/5:n] / float(n - 1)

plt.title("(a/c) / (a/c)'")
plt.xlabel("a")
plt.ylabel("b")
grid = intr.griddata(pts, ac_rat, (g_ac, g_bc), method="linear")
grid[np.isnan(grid)] = 0
grid = np.maximum(grid, grid.T)
ac_grid = grid
for i in xrange(len(grid)):
    grid[i,i] = grid[i,i-1]
plt.imshow(grid, extent=[0, 1, 0, 1], interpolation="nearest")
plt.colorbar()

plt.figure()

plt.title("(b/c) / (b/c)'")
grid = intr.griddata(pts, bc_rat, (g_ac, g_bc), method="linear")
grid[np.isnan(grid)] = 0
grid = np.maximum(grid, grid.T)
for i in xrange(len(grid)):
    grid[i,i] = grid[i,i-1]
grid[0, 0] = grid[0, 1]
bc_grid = grid
plt.imshow(grid, extent=[0, 1, 0, 1], interpolation="nearest")
plt.colorbar()

plt.figure()

plt.title("c / c'")
grid = intr.griddata(pts, c_rat, (g_ac, g_bc), method="linear")
grid[np.isnan(grid)] = 0
grid = np.maximum(grid, grid.T)
for i in xrange(len(grid)):
    grid[i,i] = grid[i,i-1]
grid[0, 0] = grid[0, 1]
c_grid = grid
plt.imshow(grid, extent=[0, 1, 0, 1], interpolation="nearest")
plt.colorbar()

def print_go_slice(name, xs):
    print "var %s = []float64{" % name
    for i in xrange(len(xs)):
        if i % 8 == 0: print "\t",
        print ("%.5g," % xs[i]),
        if i % 8 == 7: print
    if len(xs) % 8 != 0: print
    print "}"

def print_go_header():
    print "package ellipse_grid"
    print

def flip(g):
    return g[::-1, :]

print_go_header()
print_go_slice("ACRatios", g_ac[0])
print_go_slice("BCRatios", g_bc.T[0])
print_go_slice("ACCorrectionGrid", ac_grid.flatten())
print_go_slice("BCCorrectionGrid", bc_grid.flatten())
print_go_slice("CCorrectionGrid", flip(c_grid).flatten())

plt.show()
