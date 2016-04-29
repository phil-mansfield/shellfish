package env

type (
	CatalogType int
	HaloType int
	TreeType int
)

const (
	Gotetra CatalogType = iota

	Rockstar HaloType = iota
	NilHalo

	ConsistentTrees TreeType = iota
	NilTree
)

type ParticleInfo struct {
	SnapshotFormat          string
	MemoDir                 string

	SnapshotFormatMeanings  []string
	ScaleFactorFile         string
	FormatMins, FormatMaxes []int64
	SnapMin, SnapMax        int64
}

type HaloInfo struct {
	HaloDir, TreeDir string
	SnapMin, SnapMax int64
}

type Environment struct {
	Catalogs
	Halos
	MemoDir string
}

//////////////
// Catalogs //
//////////////

type Catalogs struct {
	CatalogType
	snapMin int
	names [][]string
}


func (cat *Catalogs) Blocks() int {
	return len(cat.names[0])
}

func (cat *Catalogs) ParticleCatalog(snap, block int) string {
	return cat.names[snap - cat.snapMin][block]
}

///////////
// Halos //
///////////

type Halos struct {
	HaloType
	TreeType
	snapMin int
	snapOffset int
	names []string
}

func (h *Halos) HaloCatalog(snap int) string {
	return h.names[snap - h.snapMin]
}

func (h *Halos) SnapOffset() int {
	return h.snapOffset
}

func (h *Halos) InitNilHalo(dir string, snapMin, snapMax int64) error {
	h.HaloType = NilHalo
	h.TreeType = NilTree

	return nil
}