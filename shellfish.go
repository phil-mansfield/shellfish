/*package shellfish contains code for computing the splashback shells of
halos in N-body simulations.*/
package main

import (
	"io/ioutil"
	"log"
	"os"
	"fmt"
	"strings"

	"github.com/phil-mansfield/shellfish/cmd"
)

func modeDescriptions() string {
	return "My help mode is:\n" +
	"$ ./shellfish help\n\n" +

	"My setup mode is:\n" +
	"$ ./shellfish setup ____.setup.config\n\n" +

	"My other modes are:\n" +
	"$ ./shellfish id     [flags] ____.config [____.id.config]\n" +
	"$ ./shellfish tree   [flags] ____.config [____.tree.config]\n" +
	"$ ./shellfish shell  [flags] ____.config [____.shell.config]\n" +
	"$ ./shellfish stats  [flags] ____.config [____.stats.config]\n\n"
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


	if args[1] == "setup" {
		// TODO: Implement the setup command.
		panic("NYI")
	} else if args[1] == "help" {
		// TODO: Implement the help command.
	}

	mode, ok := cmd.ModeNames[args[1]]
	if !ok {
		fmt.Fprintf(
			os.Stderr, "You passed me the mode '%s', which I don't " +
			"recognize.\nFor help, type './shellfish help'\n", args[1],
		)
		os.Exit(1)
	}

	var lines []string
	switch args[1] {
	case "tree", "shell", "stats":
		var err error
		lines, err = stdinLines()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	flags := getFlags(args)
	gConfig, err := getGlobalConfig(args)
	if err != nil {
		log.Fatalf("Error running mode %s:\n%s\n", args[1], err.Error())
	}
	config, ok := getConfig(args)

	if ok {
		if err = mode.ReadConfig(config); err != nil {
			log.Fatalf("Error running mode %s:\n%s\n", args[1], err.Error())
		}
	}

	out, err := mode.Run(flags, gConfig, lines)
	if err != nil {
		log.Fatalf("Error running mode %s:\n%s\n", args[1], err.Error())
	}

	for i := range out { fmt.Println(out[i]) }
}


// stdinLines reads stdin and splits it into lines.
func stdinLines() ([]string, error) {
	bs, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf(
			"Error reading stdin: %s.", err.Error(),
		)
	}
	text := string(bs)
	lines := strings.Split(text, "\n")
	if lines[len(lines) - 1] == "" { lines = lines[:len(lines) - 1] }
	return lines, nil
}

// getFlags reutrns the flag tokens from the command line arguments.
func getFlags(args []string) ([]string) {
	return args[1:len(args) - 1 - configNum(args)]
}

// getGlobalConfig returns the name of the base config file from the command
// line arguments.
func getGlobalConfig(args []string) (*cmd.GlobalConfig, error) {
	name := ""
	switch configNum(args) {
	case 0:
		return nil, fmt.Errorf("No config files provided in command " +
			"line arguments")
	case 1:
		name = args[len(args) - 1]
	case 2:
		name = args[len(args) - 2]
	}

	config := &cmd.GlobalConfig{}
	err := config.ReadConfig(name)
	if err != nil { return nil, err }
	return config, nil
}

// getConfig return the name of the mode-specific config file from the command
// line arguments.
func getConfig(args []string) (string, bool) {
	if configNum(args) == 2 {
		return args[len(args) - 1], true
	}
	return "", false
}

// configNum returns the number of configuration files at the end of the
// argument list (up to 2).
func configNum(args []string) int {
	num := 0
	for i := len(args) - 1; i >= 0 && i >= len(args) - 2; i-- {
		if isConfig(args[i]) {
			num++
		} else {
			break
		}
	}
	return num
}

// isConfig returns true if the fiven string is a config file name.
func isConfig(s string) bool {
	return len(s) >= 7 &&  s[len(s) - 7:] == ".config"
}