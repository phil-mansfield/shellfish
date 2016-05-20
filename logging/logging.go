package logging

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
