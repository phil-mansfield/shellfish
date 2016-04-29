package cmd

import (
	"fmt"
	"github.com/phil-mansfield/shellfish/io"
)

func getVectorBuffer(
	fname, typeString, endiannessString string,
) (io.VectorBuffer, error) {
	switch typeString {
	case "gotetra":
		return io.NewGotetraBuffer(fname)
	case "LGadget-2":
		return io.NewLGadget2Buffer(fname, endiannessString)
	case "ARTIO":
		return io.NewARTIOBuffer(fname)
	}
	// Impossible, but worth doing anyway.
	return nil, fmt.Errorf("SnapshotType '%s' not recognized.", typeString)
}
