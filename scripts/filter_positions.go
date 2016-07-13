package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	
	"github.com/phil-mansfield/shellfish/io"
)

const (
	L = 250.0
	N = 1024
	G = 8
	BaseDir = "/project/surph/mansfield/data/sheet_segments/Box_L0250_N1024_G0008_CBol/snapdir_%03d/"
)

// TODO: Rewrite this monstrosity so that it only reads a single file in once.

func read(snap int, ids []int64) [][3]float32 {
	dir := fmt.Sprintf(BaseDir, snap)

	buf, err := io.NewGotetraBuffer(path.Join(dir, "sheet000.dat"))
	if err != nil { panic(err.Error()) }
	
	Gn := int64(N / G)

	var vecs [][3]float32
	out := [][3]float32{}
	
	Gis := make([]int64, len(ids))
	flags := make([]bool, 8*8*8)
	for idx, id := range ids {
		ix, iy, iz := id % N, (id/N) % N, id / (N*N)
		// I don't even.
		ix -= 1
		if ix == -1 { ix = 1023 }
		
		Gx, Gy, Gz := ix / Gn, iy / Gn, iz / Gn
		Gi := Gx + Gy*G + Gz*G*G

		Gis[idx] = Gi
		flags[Gi] = true
	}
	for Gi := int64(0); Gi < int64(len(flags)); Gi++ {
		if !flags[Gi] { continue }
		runtime.GC()

		Gx, Gy, Gz := int64(Gi) % 8, (int64(Gi) / 8) % 8, int64(Gi) / (8*8)
		vecs, _, _, err = buf.Read(
			path.Join(dir, fmt.Sprintf("sheet%d%d%d.dat", Gx, Gy, Gz)),
		)
		if err != nil {
			panic(err.Error())
		}

		for idx, id := range ids {
			if Gis[idx] != Gi { continue }

			ix, iy, iz := id % N, (id/N) % N, id / (N*N)
			ix -= 1
			if ix == -1 { ix = 1023 }
			jx, jy, jz := ix - Gx*Gn, iy - Gy*Gn, iz - Gz*Gn
			if jx < 0 || jy < 0 || jz < 0 || jx >= 128 || jy >= 128 || jz >= 128 {
				panic(":3")
			}
			out = append(out, vecs[jx + jy*Gn + jz*Gn*Gn])
		}

		buf.Close()
	}
	
	return out
}

func vecAlmostEq(v1, v2 [3]float32, eps float32) bool {
	return almostEq(v1[0], v2[0], eps) &&  almostEq(v1[1], v2[1], eps) &&
		almostEq(v1[2], v2[2], eps)
}

func almostEq(x, y, eps float32) bool {
	return x + eps > y && x - eps < y
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Correct usage: '$ filter_positions filter_data.dat start_halo'")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil { panic(err.Error()) }
	defer f.Close()
	
	data, err := io.ReadFilter(f)

	startHalo, err := strconv.Atoi(os.Args[2])
	if err != nil { panic(err.Error()) }
	for h := startHalo; h < len(data.Snaps); h++ {
		fmt.Printf("%d/%d (%d)\n", h, len(data.Snaps), len(data.Particles[h]))
		for _, snap := range []int{ 100, 99, 98, 97, 96, 95, 94, 93, 92, 91, 90 } {
			fmt.Println("   ", snap)
			if err != nil { panic(err.Error()) }
			
			os.MkdirAll(fmt.Sprintf("data/h%d", h), os.ModeDir | os.ModePerm)
			f, err = os.Create(fmt.Sprintf("data/h%d/shell_p_%d.dat", h, snap))
			defer f.Close()
			vecs := read(snap, data.Particles[h])
			
			for _, vec := range vecs {
				x, y, z := vec[0], vec[1], vec[2]
				fmt.Fprintf(f, "%7.6g %7.6g %7.6g\n", x, y, z)
			}
		}
	}
}
