package env

import (
	"fmt"
)


func (cat *Catalogs) InitNil(info *ParticleInfo, validate bool) error {
	cat.CatalogType = Nil
	cat.names = make([][]string, info.SnapMax - info.SnapMin + 1)

	for i := range cat.names { cat.names[i] = []string{fmt.Sprintf("%d", i)} }

	if validate {
		panic("File validation not yet implemented.")
	}
	return nil
}
