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

// MemString returns a string containing various statistics on the current
// memory usage of Shellfish.
func MemString() string {
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	return fmt.Sprintf(
		"Alloc - %d MB; Sys - %d MB Integrated - %d MB",
		ms.Alloc >> 20, ms.Sys >> 20, ms.TotalAlloc >> 20,
	)
}
