package logging

import (
	"fmt"
	"runtime"
)

type Flag int

const (
	Nil Flag = iota
	Performance
	Debug
)

// This is handled this way so that GlobalConfig doesn't need to be literally
// every function in the project.
var (
	Mode Flag = Nil
)

func MemString() string {
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	return fmt.Sprintf(
		`Program Allocation: %d MB
System Allocation: %d MB
Integrated Allocation: %d MB`, ms.Alloc, ms.Sys, ms.TotalAlloc,
	)
}
