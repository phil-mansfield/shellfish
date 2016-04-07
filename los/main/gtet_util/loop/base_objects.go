package loop

// Sphere is a struct that can be used as an embedded struct in structs which
// are implementing the Object interface and are shaped like a sphere.
type Sphere struct {
	origin [3]float64
	rMin, rMax, rMin2, rMax2, tw float64
}

// Init initrializes a Sphere.
func (s *Sphere) Init(origin [3]float64, rMin, rMax float64) {
	s.rMin = rMin
	s.rMax = rMax
	s.rMin2 = rMin * rMin
	s.rMax2 = rMax * rMax
	s.origin = origin
}

// Transform does a coordinate transformation on a slice of vectors so that they
// are as close to the center of the Sphere.
func (s *Sphere) Transform(vecs [][3]float32, totalWidth float64) {
    x0 := float32(s.origin[0])
    y0 := float32(s.origin[1])
    z0 := float32(s.origin[2])
    tw := float32(totalWidth)
    tw2 := tw / 2

    for i, vec := range vecs {
		x, y, z := vec[0], vec[1], vec[2]
		dx, dy, dz := x - x0, y - y0, z - z0

        if dx > tw2 {
            vecs[i][0] -= tw
        } else if dx < -tw2 {
            vecs[i][0] += tw
        }

        if dy > tw2 {
            vecs[i][1] -= tw
        } else if dy < -tw2 {
            vecs[i][1] += tw
        }

        if dz > tw2 {
            vecs[i][2] -= tw
        } else if dz < -tw2 {
            vecs[i][2] += tw
        }
	}
}

// Contains returns true if a point is inside the Sphere
func (s *Sphere) Contains(x, y, z float64) bool {
    x0, y0, z0 := s.origin[0], s.origin[1], s.origin[2]
    dx, dy, dz := x - x0, y - y0, z - z0
    r2 :=  dx*dx + dy*dy + dz*dz
	return s.rMin2 < r2 && r2 < s.rMax2
}

// IntersectBox returns true if the Sphere interects a box.
func (s *Sphere) IntersectBox(origin, span [3]float64, tw float64) bool {
	return inRange(s.origin[0], s.rMax, origin[0], span[0], tw) &&
		inRange(s.origin[1], s.rMax, origin[1], span[1], tw) &&
		inRange(s.origin[2], s.rMax, origin[2], span[2], tw)
}

func inRange(x, r, low, width, tw float64) bool {
	return wrapDist(x, low, tw) > -r && wrapDist(x, low + width, tw) < r
}

func wrapDist(x1, x2, width float64) float64 {
	dist := x1 - x2
	if dist > width / 2 {
		return dist - width
	} else if dist < width / -2 {
		return dist + width
	} else {
		return dist
	}
}

// Box is a struct that can be used as an embedded struct in structs which
// are implementing the Object interface and are shaped like a box.
type Box struct {
	origin, span [3]float64
}

// Init initializes a Box
func (b *Box) Init(origin, span [3]float64) {
	b.origin = origin
	b.span = span
}

// Transform does a coordinate transformation on a slice of vectors so that they
// are as close to the center of the Box.
func (b *Box) Transform(vecs [][3]float32, totalWidth float64) {
    x0 := float32(b.origin[0])
    y0 := float32(b.origin[1])
    z0 := float32(b.origin[2])
    tw := float32(totalWidth)
    tw2 := tw / 2

    for i, vec := range vecs {
		x, y, z := vec[0], vec[1], vec[2]
		dx, dy, dz := x - x0, y - y0, z - z0

        if dx > tw2 {
            vecs[i][0] -= tw
        } else if dx < -tw2 {
            vecs[i][0] += tw
        }

        if dy > tw2 {
            vecs[i][1] -= tw
        } else if dy < -tw2 {
            vecs[i][1] += tw
        }

        if dz > tw2 {
            vecs[i][2] -= tw
        } else if dz < -tw2 {
            vecs[i][2] += tw
        }
	}
}

// Contains returns true if a point is inside the Box.
func (b *Box) Contains(x, y, z float64) bool {
	lowX, highX := b.origin[0], b.origin[0] + b.span[0]
	lowY, highY := b.origin[1], b.origin[1] + b.span[1]
	lowZ, highZ := b.origin[2], b.origin[2] + b.span[2]
	return lowX < x && x < highX && 
		lowY < y && y < highY && 
		lowZ < z && z < highZ
}

// IntersectBox returns true if the Box intersects with another Box.
func (b *Box) IntersectBox(origin, span [3]float64, tw float64) bool {
	s2x := b.span[0] / 2
	s2y := b.span[1] / 2
	s2z := b.span[2] / 2

	return inRange(b.origin[0] + s2x, s2x, origin[0], span[0], tw) &&
		inRange(b.origin[1] + s2y, s2y, origin[1], span[1], tw) &&
		inRange(b.origin[2] + s2z, s2z, origin[2], span[2], tw)
}
