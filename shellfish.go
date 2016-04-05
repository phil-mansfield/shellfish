/*package shellfish contains code for computing the splashback shells of
halos in N-body simulations.*/
package main

import (
	"os"
	"fmt"
)

func modeDescriptions() string {
	return "My help mode is:\n" +
	"$ ./shellfish help\n\n" +

	"My setup mode is:\n" +
	"$ ./shellfish setup ____.setup.config\n\n" +

	"My shell-finding modes are:\n" +
	"$ ./shellfish id    [flags] ____.config [____.id.config]\n" +
	"$ ./shellfish tree  [flags] ____.config [____.tree.config]\n" +
	"$ ./shellfish shell [flags] ____.config [____.shell.config]\n" +
	"$ ./shellfish stats [flags] ____.config [____.stats.config]\n\n" +

	"I have a few other modes which are convenient for this type of " +
	"analysis:\n" +
	"$ ./shellfish append [flags] ____.config [____.append.config]\n" +
	"$ ./shellfish prof   [flags] ____.config [____.id.config]\n" +
	"$ ./shellfish render [flags] ____.config [____.render.config]\n" +
	"$ ./shellfish phase  [flags] ____.config [____.phase]\n"
}

func main() {
	args := os.Args
	if len(args) <= 1 {
		fmt.Fprintf(
			os.Stderr, "I was not supplied with a mode.\nFor help, type " +
			"'./shellfish help'.\n",
		)
		os.Exit(1)
	}

	switch args[1] {
	case "":
	default:
		fmt.Fprintf(
			os.Stderr, "You passed me the mode '%s', which I don't " +
			"recognize.\nFor help, type './shellfish help'\n", args[1],
		)
		os.Exit(1)
	}
}
