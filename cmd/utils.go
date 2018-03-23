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
		NilTotalWidth: config.NilSnapTotalWidth,
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
	case "BolshoiP":
		return io.NewBolshoiPBuffer(fname, config.Endianness, context)
	case "nil":
		return io.NewNilBuffer(context)
	}

	// Impossible, but worth doing anyway.
	return nil, fmt.Errorf(
		"SnapshotType '%s' not recognized.", config.SnapshotType,
	)
}

// How to use:
//
// lg := NewLockGroup(workers)
// for i := 0; i < workers; i++ {
//     go f(lg.Lock(i))
// }
//
// lg.Synchronize()
//
// func f(l *Lock) {
//     ... do work, decided by l.Workers and l.Idx ...
//     l.Unlock()
// }

type Lock struct {
	Workers, Idx int
	C chan bool
}

func (l *Lock) Unlock() {
	l.C <- true
}

type LockGroup struct {
	Workers int
	C chan bool
	locks []Lock
}

func NewLockGroup(workers int) *LockGroup {
	lg := &LockGroup{}
	lg.C = make(chan bool, workers)
	lg.Workers = workers
	lg.locks = make([]Lock, workers)
	for i := range lg.locks {
		lg.locks[i].Workers = workers
		lg.locks[i].Idx = i
		lg.locks[i].C = lg.C
	}

	return lg
}

func (lg *LockGroup) Lock(i int) *Lock {
	return &lg.locks[i]
}

func (lg * LockGroup) Synchronize() {
	for _ = range lg.locks {
		<- lg.C
	}
}