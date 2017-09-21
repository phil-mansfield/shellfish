package cmd

import (
	"fmt"
	"github.com/phil-mansfield/shellfish/io"
)

func getVectorBuffer(
	fname string, config *GlobalConfig,
) (io.VectorBuffer, error) {
	context := io.Context{
		LGadgetNPartNum: config.LGadgetNpartNum,
		GadgetDMTypeIndices: config.GadgetDMTypeIndices,
		GadgetDMSingleMassIndices: config.GadgetSingleMassIndices,
		GadgetMassUnits: config.GadgetMassUnits,
		GadgetPositionUnits: config.GadgetPositionUnits,
		NilOmegaM: config.NilSnapOmegaM,
		NilOmegaL: config.NilSnapOmegaL,
		NilH100: config.NilSnapH100,
		NilScaleFactors: config.NilSnapScaleFactors,
	}
	
	switch config.SnapshotType {
	case "gotetra":
		return io.NewGotetraBuffer(fname)
	case "LGadget-2":
		return io.NewLGadget2Buffer(fname, config.Endianness, context)
	case "Gadget-2":
		return io.NewGadget2Buffer(fname, config.Endianness, context)
	case "ARTIO":
		return io.NewARTIOBuffer(fname)
	case "Bolshoi":
		return io.NewBolshoiBuffer(fname, config.Endianness, context)
	case "nil":
		return io.NewNilBuffer(context)
	}

	// Impossible, but worth doing anyway.
	return nil, fmt.Errorf(
		"SnapshotType '%s' not recognized.", config.SnapshotType,
	)
}
