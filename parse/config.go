package parse

import (
	"io/ioutil"
	"strconv"
	"strings"
	"fmt"
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
	case intVar: return "int"
	case intsVar: return "int list"
	case floatVar: return "float"
	case floatsVar: return "float list"
	case stringVar: return "string"
	case stringsVar: return "string list"
	case boolVar: return "bool"
	case boolsVar: return "bool list"
	}
	panic("Impossible")
}

type conversionFunc func(string) bool

type ConfigVars struct {
	name string
	varNames []string
	varTypes []varType
	conversionFuncs []conversionFunc
}

func intConv(ptr *int64) conversionFunc {
	return func(s string) bool {
		i, err := strconv.Atoi(s)
		if err != nil { return false }
		*ptr = int64(i)
		return true
	}
}

func floatConv(ptr *float64) conversionFunc {
	return func(s string) bool {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil { return false }
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
		if err != nil { return false }
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
		*ptr = (*ptr)[:0]
		for j := range toks {
			i, err := strconv.Atoi(toks[j])
			if err != nil { return false }
			*ptr = append(*ptr, int64(i))
		}
		return true
	}
}

func floatsConv(ptr *[]float64) conversionFunc {
	return func(s string) bool {
		toks := strToList(s)
		for j := range toks {
			f, err := strconv.ParseFloat(toks[j], 64)
			if err != nil { return false }
			*ptr = append(*ptr, f)
		}
		return true
	}
}

func stringsConv(ptr *[]string) conversionFunc {
	return func(s string) bool {
		toks := strToList(s)
		for j := range toks {
			*ptr = append(*ptr, toks[j])
		}
		return true
	}
}


func boolsConv(ptr *[]bool) conversionFunc {
	return func(s string) bool {
		toks := strToList(s)
		for j := range toks {
			b, err := strconv.ParseBool(toks[j])
			if err != nil { return false }
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

func ReadConfig(fname string, vars *ConfigVars) error {
	for i := range vars.varNames {
		vars.varNames[i] = strings.ToLower(vars.varNames[i])
	}

	bs, err := ioutil.ReadFile(fname)
	if err != nil { return err }
	lines := strings.Split(string(bs), "\n")
	lines, lineNums := removeComments(lines)
	for i := range lineNums { lineNums[i] ++ }

	if len(lines) == 0 || lines[0] != fmt.Sprintf("[%s]", vars.name) {
		return fmt.Errorf(
			"I expected the config file %s to have the header " +
			"[%s] at the top, but didn't find it.", fname, vars.name,
		)
	}
	lines = lines[1:]

	names, vals, errLine := associationList(lines)
	if errLine !=  -1 {
		return fmt.Errorf(
			"I could not parse line %d of the config file %s because it " +
			"did not take the form of a variable assignment.",
			lineNums[errLine+1], fname,
		)
	}

	if errLine = checkValidNames(names, vars); errLine != -1 {
		return fmt.Errorf(
			"Line %d of the config file %s assigns a value to the " +
			"variable '%s', but config files of type %s don't have that " +
			"variable.", lineNums[errLine+1], fname, names[errLine], vars.name,
		)
	}

	if errLine1, errLine2 := checkDuplicateNames(names); errLine1 != -1 {
		return fmt.Errorf(
			"Lines %d and %d of the config file %s both assign a value to " +
			"the variable '%s'.", lineNums[errLine1+1], lineNums[errLine2+1],
			fname, names[errLine1],
		)
	}

	if errLine = convertAssoc(names, vals, vars); errLine != -1 {
		j := 0
		for ; j < len(vars.varNames); j++ {
			if vars.varNames[j] == names[errLine] { break }
		}
		typeName := vars.varTypes[j].String()
		a := "a"
		if typeName[0] == 'i' { a = "an" }
		return fmt.Errorf(
			"I could not parse line %d of the config file %s because '%s' " +
			"expects values of type %s and '%s' cannnot be converted to " +
			"%s %s.", lineNums[errLine+1], fname, vars.varNames[j], typeName,
			vals[j], a, typeName,
		)
	}

	return nil
}

func removeComments(lines []string) ([]string, []int) {
	tmp := make([]string, len(lines))
	copy(tmp, lines)
	lines = tmp

	for i := range lines {
		comment := strings.Index(lines[i], "#")
		if comment == -1 { continue }
		lines[i] = lines[i][:comment]
	}

	out, lineNums := []string{}, []int{}
	for i := range lines {
		line := strings.Trim(lines[i], " ")
		if len(line) == 0 { continue }
		out = append(out, line)
		lineNums = append(lineNums, i)
	}

	return out, lineNums
}

func associationList(lines []string) ([]string, []string, int) {
	names, vals := []string{}, []string{}
	for i := range lines {
		eq := strings.Index(lines[i], "=")
		if eq == -1 { return nil, nil, i }
		name := lines[i][:eq]
		val := ""
		if len(lines[i]) - 1 > eq { val = lines[i][eq+1:] }
		names = append(names, strings.ToLower(strings.Trim(name, " ")))
		if len(names[len(names) - 1]) == 0 { return nil, nil, i }
		vals = append(vals, strings.Trim(val, " "))
	}
	return names, vals, -1
}

func checkValidNames(names []string, vars *ConfigVars) int {
	for i := range names {
		found := false
		for j := range vars.varNames {
			if vars.varNames[j] == names[i] {
				found = true
				break
			}
		}
		if !found { return i }
	}
	return -1
}

func checkDuplicateNames(names []string) (int, int) {
	for i := range names {
		for j := i + 1; j < len(names); j++ {
			if names[i] == names[j] { return i, j }
		}
	}
	return -1, -1
}

func convertAssoc(names, vals []string, vars *ConfigVars) int {
	for i := range names {
		j := 0
		for ; j < len(vars.varNames); j++ {
			if vars.varNames[j] == names[i] { break }
		}

		ok := vars.conversionFuncs[j](vals[i])
		if !ok { return i }
	}
	return -1
}