/*package parse contains routines for parsing config files. To parse a config
file, the user creates a ConfigVars struct and registers every variable in the
config file with it. Registration requires 1) the variable name, 2) a
default value, and 3) a location to write the variable to.

Config files have two parse with this package have two parts, a title and a
body. The title specifies the type of config file an the body contains
Variable = Value pairs. The title ensures that when a project has multiple
config files the wrong one isn't read by mistake.

Here is an example configuration file that collects information about
my cat:

    # Title
    [cat_info]

    # Body:
	CatName = Bob
	FurColors = White, Black
	Age = 7.5 # Inline comments are okay, too.
	Paws = 4

Here is an example of using the parse package to parse this type of config
file.

	type CatInfo struct {
		CatName string
		FurColors []string
		Age float
		Paws, Tails int
	}

	info := new(CatInfo)

    vars := ConfigVars("cat_info")
    vars.String(&info.CatName, "CatName", "")
    vars.Strings(&info.FurColors, "FurColors", []string{})
    vars.Float(&info.Age, "Age", -1)
    vars.Int(&info.Paws, "Paws", 4)
    vars.Int(&info.Tail, "Tail", 1)

    // Then, once a file has been provided

    err := ReadConfig("my_cat.config", vars)
    if err != nil {
        // Handle error
    }

A careful read of the above example will show that the supplied config file
does not consider the config file missing one or more variables an error. This
will be annoying in some cases, but is usually the desired behavior. You will
need to explicitly check for variables that have not been set.

For additional examples, see the usage in config_test.go
*/
package parse

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

/////////////////////
// Conversion Code //
/////////////////////

type varType int

const (
	intVar varType = iota
	intsVar
	floatVar
	floatsVar
	stringVar
	stringsVar
	boolVar
	boolsVar
)

func (v varType) String() string {
	switch v {
	case intVar:
		return "int"
	case intsVar:
		return "int list"
	case floatVar:
		return "float"
	case floatsVar:
		return "float list"
	case stringVar:
		return "string"
	case stringsVar:
		return "string list"
	case boolVar:
		return "bool"
	case boolsVar:
		return "bool list"
	}
	panic("Impossible")
}

type conversionFunc func(string) bool

type ConfigVars struct {
	name            string
	varNames        []string
	varTypes        []varType
	conversionFuncs []conversionFunc
}

func intConv(ptr *int64) conversionFunc {
	return func(s string) bool {
		i, err := strconv.Atoi(s)
		if err != nil {
			return false
		}
		*ptr = int64(i)
		return true
	}
}

func floatConv(ptr *float64) conversionFunc {
	return func(s string) bool {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return false
		}
		*ptr = f
		return true
	}
}

func stringConv(ptr *string) conversionFunc {
	return func(s string) bool {
		*ptr = strings.Trim(s, " ")
		return true
	}
}

func boolConv(ptr *bool) conversionFunc {
	return func(s string) bool {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return false
		}
		*ptr = b
		return true
	}
}

func strToList(a string) []string {
	strs := strings.Split(a, ",")
	for i := range strs {
		strs[i] = strings.Trim(strs[i], " ")
	}
	return strs
}

func intsConv(ptr *[]int64) conversionFunc {
	return func(s string) bool {
		toks := strToList(s)
		*ptr = []int64{}
		for j := range toks {
			i, err := strconv.Atoi(toks[j])
			if err != nil {
				return false
			}
			*ptr = append(*ptr, int64(i))
		}
		return true
	}
}

func floatsConv(ptr *[]float64) conversionFunc {
	return func(s string) bool {
		toks := strToList(s)
		*ptr = []float64{}
		for j := range toks {
			f, err := strconv.ParseFloat(toks[j], 64)
			if err != nil {
				return false
			}
			*ptr = append(*ptr, f)
		}
		return true
	}
}

func stringsConv(ptr *[]string) conversionFunc {
	return func(s string) bool {
		toks := strToList(s)
		*ptr = []string{}
		for j := range toks {
			*ptr = append(*ptr, toks[j])
		}
		return true
	}
}

func boolsConv(ptr *[]bool) conversionFunc {
	return func(s string) bool {
		toks := strToList(s)
		*ptr = []bool{}
		for j := range toks {
			b, err := strconv.ParseBool(toks[j])
			if err != nil {
				return false
			}
			*ptr = append(*ptr, b)
		}
		return true
	}
}

func NewConfigVars(name string) *ConfigVars {
	return &ConfigVars{name: name}
}

func (vars *ConfigVars) Int(ptr *int64, name string, value int64) {
	*ptr = value
	vars.varNames = append(vars.varNames, name)
	vars.conversionFuncs = append(vars.conversionFuncs, intConv(ptr))
	vars.varTypes = append(vars.varTypes, intVar)
}

func (vars *ConfigVars) Float(ptr *float64, name string, value float64) {
	*ptr = value
	vars.varNames = append(vars.varNames, name)
	vars.conversionFuncs = append(vars.conversionFuncs, floatConv(ptr))
	vars.varTypes = append(vars.varTypes, floatVar)
}

func (vars *ConfigVars) String(ptr *string, name string, value string) {
	*ptr = value
	vars.varNames = append(vars.varNames, name)
	vars.conversionFuncs = append(vars.conversionFuncs, stringConv(ptr))
	vars.varTypes = append(vars.varTypes, stringVar)
}

func (vars *ConfigVars) Bool(ptr *bool, name string, value bool) {
	*ptr = value
	vars.varNames = append(vars.varNames, name)
	vars.conversionFuncs = append(vars.conversionFuncs, boolConv(ptr))
	vars.varTypes = append(vars.varTypes, boolVar)
}

func (vars *ConfigVars) Ints(ptr *[]int64, name string, value []int64) {
	*ptr = value
	vars.varNames = append(vars.varNames, name)
	vars.conversionFuncs = append(vars.conversionFuncs, intsConv(ptr))
	vars.varTypes = append(vars.varTypes, intsVar)
}

func (vars *ConfigVars) Floats(ptr *[]float64, name string, value []float64) {
	*ptr = value
	vars.varNames = append(vars.varNames, name)
	vars.conversionFuncs = append(vars.conversionFuncs, floatsConv(ptr))
	vars.varTypes = append(vars.varTypes, floatsVar)
}

func (vars *ConfigVars) Strings(ptr *[]string, name string, value []string) {
	*ptr = value
	vars.varNames = append(vars.varNames, name)
	vars.conversionFuncs = append(vars.conversionFuncs, stringsConv(ptr))
	vars.varTypes = append(vars.varTypes, stringsVar)
}

func (vars *ConfigVars) Bools(ptr *[]bool, name string, value []bool) {
	*ptr = value
	vars.varNames = append(vars.varNames, name)
	vars.conversionFuncs = append(vars.conversionFuncs, boolsConv(ptr))
	vars.varTypes = append(vars.varTypes, boolsVar)
}

//////////////////
// Parsing Code //
//////////////////

// ReadConfig parses the config file specified by fname using the set of
// variables vars. If successful nil is returned, otherwise an error is
// returned.
func ReadConfig(fname string, vars *ConfigVars) error {
	// I/O

	bs, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	// Begin tokenization; remember line numbers for better errors.

	lines := strings.Split(string(bs), "\n")
	lines, lineNums := removeComments(lines)
	for i := range lineNums {
		lineNums[i]++
	}

	if len(lines) == 0 || lines[0] != fmt.Sprintf("[%s]", vars.name) {
		return fmt.Errorf(
			"I expected the config file %s to have the header "+
				"[%s] at the top, but didn't find it.", fname, vars.name,
		)
	}
	lines = lines[1:]

	// Create association list and check for name-based errors

	names, vals, errLine := associationList(lines)
	if errLine != -1 {
		return fmt.Errorf(
			"I could not parse line %d of the config file %s because it "+
				"did not take the form of a variable assignment.",
			lineNums[errLine+1], fname,
		)
	}

	if errLine = checkValidNames(names, vars); errLine != -1 {
		return fmt.Errorf(
			"Line %d of the config file %s assigns a value to the "+
				"variable '%s', but config files of type %s don't have that "+
				"variable.", lineNums[errLine+1], fname, names[errLine], vars.name,
		)
	}

	if errLine1, errLine2 := checkDuplicateNames(names); errLine1 != -1 {
		return fmt.Errorf(
			"Lines %d and %d of the config file %s both assign a value to "+
				"the variable '%s'.", lineNums[errLine1+1], lineNums[errLine2+1],
			fname, names[errLine1],
		)
	}

	// Convert every variable in the associate list.

	if errLine = convertAssoc(names, vals, vars); errLine != -1 {
		j := 0
		for ; j < len(vars.varNames); j++ {
			if strings.ToLower(vars.varNames[j]) ==
				strings.ToLower(names[errLine]) {
				break
			}
		}
		typeName := vars.varTypes[j].String()
		a := "a"
		if typeName[0] == 'i' {
			a = "an"
		}
		return fmt.Errorf(
			"I could not parse line %d of the config file %s because '%s' "+
				"expects values of type %s and '%s' cannnot be converted to "+
				"%s %s.", lineNums[errLine+1], fname, vars.varNames[j], typeName,
			vals[j], a, typeName,
		)
	}

	return nil
}

func ReadFlags(args []string, vars *ConfigVars) error {
	if len(args) == 0 { return nil }
	for _, arg := range args {
		for j := range arg {
			if arg[j] == '=' {
				return fmt.Errorf(
					"The argument '%s' contains an equals sign.", arg,
				)
			}
		}
	}

	isFlag := make([]bool, len(args))
	for i := range args {
		isFlag[i] = len(args[i]) > 1 && args[i][:2] == "--"
	}

	if !isFlag[0] {
		return fmt.Errorf("The argument '%s' does not have a flag.", args[0])
	}

	varNames, values := []string{}, []string{}
	currValue := []string{}

	varNames = append(varNames, flagVarName(args[0]))
	for i := 1; i < len(args); i++ {
		if !isFlag[i] {
			currValue = append(currValue, args[i])
		} else {
			valStr := strings.Join(currValue, ",")
			values = append(values, valStr)
			currValue = []string{}
			varNames = append(varNames, flagVarName(args[i]))
		}
	}
	valStr := strings.Join(currValue, ",")
	values = append(values, valStr)
	
	for i, value := range values {
		if value == "" {
			return fmt.Errorf(
				"The flag '%s' was supplied, but wasn't set to a value.",
				varNames[i],
			)
		}
	}

	lines := make([]string, len(values))
	for i := range lines {
		lines[i] = fmt.Sprintf("%s=%s", varNames[i], values[i])
	}

	// From here on, we're almost identical to ReadConfig
	names, vals, errLine := associationList(lines)
	if errLine != -1 {
		panic(fmt.Sprintf(
			"Internal error! Flag %d could not be parsed. " +
			"Please report this to Shellfish's current maintainer.", errLine,
		))
	}

	if errLine = checkValidNames(names, vars); errLine != -1 {
		return fmt.Errorf(
			"The flag '%s' cannot be set for this program.", names[errLine],
		)
	}

	if errLine1, _ := checkDuplicateNames(names); errLine1 != -1 {
		return fmt.Errorf(
			"The flag '%s' was assigned twice.", names[errLine1],
		)
	}

	// Convert every variable in the associate list.
	if errLine = convertAssoc(names, vals, vars); errLine != -1 {
		j := 0
		for ; j < len(vars.varNames); j++ {
			if strings.ToLower(vars.varNames[j]) ==
				strings.ToLower(names[errLine]) {
				break
			}
		}
		typeName := vars.varTypes[j].String()
		a := "a"
		if typeName[0] == 'i' {
			a = "an"
		}
		return fmt.Errorf(
			"I could not parse the flag '%s', because it "+
			"expects values of type %s and '%s' cannnot be converted to "+
			"%s %s.", vars.varNames[j], typeName, vals[j], a, typeName,
		)
	}

	return nil
}

func flagVarName(flag string) string {
	return strings.TrimLeft(flag, "-")
}

// These functions are self-explanatory.

func removeComments(lines []string) ([]string, []int) {
	tmp := make([]string, len(lines))
	copy(tmp, lines)
	lines = tmp

	for i := range lines {
		comment := strings.Index(lines[i], "#")
		if comment == -1 {
			continue
		}
		lines[i] = lines[i][:comment]
	}

	out, lineNums := []string{}, []int{}
	for i := range lines {
		line := strings.Trim(lines[i], " ")
		if len(line) == 0 {
			continue
		}
		out = append(out, line)
		lineNums = append(lineNums, i)
	}

	return out, lineNums
}

func associationList(lines []string) ([]string, []string, int) {
	names, vals := []string{}, []string{}
	for i := range lines {
		eq := strings.Index(lines[i], "=")
		if eq == -1 {
			return nil, nil, i
		}
		name := lines[i][:eq]
		val := ""
		if len(lines[i])-1 > eq {
			val = lines[i][eq+1:]
		}
		names = append(names, strings.Trim(name, " "))
		if len(names[len(names)-1]) == 0 {
			return nil, nil, i
		}
		vals = append(vals, strings.Trim(val, " "))
	}
	return names, vals, -1
}

func checkValidNames(names []string, vars *ConfigVars) int {
	for i := range names {
		found := false
		for j := range vars.varNames {
			if strings.ToLower(vars.varNames[j]) ==
				strings.ToLower(names[i]) {
				found = true
				break
			}
		}
		if !found {
			return i
		}
	}
	return -1
}

func checkDuplicateNames(names []string) (int, int) {
	for i := range names {
		for j := i + 1; j < len(names); j++ {
			if strings.ToLower(names[i]) == strings.ToLower(names[j]) {
				return i, j
			}
		}
	}
	return -1, -1
}

func convertAssoc(names, vals []string, vars *ConfigVars) int {
	for i := range names {
		j := 0
		for ; j < len(vars.varNames); j++ {
			if strings.ToLower(vars.varNames[j]) == strings.ToLower(names[i]) {
				break
			}
		}

		ok := vars.conversionFuncs[j](vals[i])
		if !ok {
			return i
		}
	}
	return -1
}
