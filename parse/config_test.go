package parse

import (
	"fmt"
	"math"
	"testing"
)

func TestIntConv(t *testing.T) {
	var x int64
	ok := intConv(&x)("41891")
	if !ok {
		t.Errorf("intConv unsuccessful on valid input.")
	}
	if x != 41891 {
		t.Errorf("intConv did not write input to pointer.")
	}
	ok = intConv(&x)("meow")
	if ok {
		t.Errorf("intConv successful on invalid input.")
	}
}

func TestFloatConv(t *testing.T) {
	var x float64
	ok := floatConv(&x)("41891.0")
	if !ok {
		t.Errorf("floatConv unsuccessful on valid input.")
	}
	if x != 41891.0 {
		t.Errorf("floatConv did not write input to pointer.")
	}
	ok = floatConv(&x)("meow")
	if ok {
		t.Errorf("floatConv successful on invalid input.")
	}
}

func TestStringConv(t *testing.T) {
	var x string
	ok := stringConv(&x)("  41891")
	if !ok {
		t.Errorf("stringConv unsuccessful on valid input.")
	}
	if x != "41891" {
		t.Errorf("stringConv did not write input to pointer.")
	}
}

func TestBoolConv(t *testing.T) {
	var x bool
	ok := boolConv(&x)("true")
	if !ok {
		t.Errorf("boolConv unsuccessful on valid input.")
	}
	if x != true {
		t.Errorf("boolConv did not write input to pointer.")
	}
	ok = boolConv(&x)("meow")
	if ok {
		t.Errorf("boolConv successful on invalid input.")
	}
}

func TestIntsConv(t *testing.T) {
	var x []int64
	ok := intsConv(&x)("1, 2 , 3")
	if !ok {
		t.Errorf("intsConv unsuccesful on valid input.")
	}
	if len(x) != 3 || x[0] != 1 || x[1] != 2 || x[2] != 3 {
		t.Errorf("intsConv did not write input to pointer.")
	}
	ok = intsConv(&x)("1,meow,3")
	if ok {
		t.Errorf("intsConv successful on invalid input.")
	}
}

func TestFloatsConv(t *testing.T) {
	var x []float64
	ok := floatsConv(&x)("1, 2.5 , 3")
	if !ok {
		t.Errorf("floatsConv unsuccesful on valid input.")
	}
	if len(x) != 3 || x[0] != 1 || x[1] != 2.5 || x[2] != 3 {
		t.Errorf("floatsConv did not write input to pointer.")
	}
	ok = floatsConv(&x)("1,meow,3")
	if ok {
		t.Errorf("floatsConv successful on invalid input.")
	}
}

func TestStringsConv(t *testing.T) {
	var x []string
	ok := stringsConv(&x)("dorothy, maddy , sahil")
	if !ok {
		t.Errorf("intsConv unsuccesful on valid input.")
	}
	if len(x) != 3 || x[0] != "dorothy" || x[1] != "maddy" || x[2] != "sahil" {
		t.Errorf("intsConv did not write input to pointer.")
	}
}

func TestBoolsConv(t *testing.T) {
	var x []bool
	ok := boolsConv(&x)("true, false,    true")
	if !ok {
		t.Errorf("intsConv unsuccesful on valid input.")
	}
	if len(x) != 3 || x[0] != true || x[1] != false || x[2] != true {
		t.Errorf("intsConv did not write input to pointer.")
	}
	ok = boolsConv(&x)("true,meow,false")
	if ok {
		t.Errorf("intsConv successful on invalid input.")
	}
}

func stringsEq(xs, ys []string) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if xs[i] != ys[i] {
			return false
		}
	}
	return true
}

func intsEq(xs, ys []int) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if xs[i] != ys[i] {
			return false
		}
	}
	return true
}

func TestRemoveComments(t *testing.T) {
	table := []struct {
		in, out  []string
		lineNums []int
	}{
		{[]string{}, []string{}, []int{}},
		{[]string{"meow"}, []string{"meow"}, []int{0}},
		{[]string{"#meow"}, []string{}, []int{}},
		{[]string{"meow", " # comment", "", "   mew "},
			[]string{"meow", "mew"}, []int{0, 3}},
	}

	for i := range table {
		res, lineNums := removeComments(table[i].in)
		if !stringsEq(table[i].out, res) {
			t.Errorf("%d) Called removeComments(%v), got %v",
				i+1, table[i].in, res)
		}
		if !intsEq(table[i].lineNums, lineNums) {
			t.Errorf("%d) Called removeComments(%v), got %v linenNums",
				i+1, table[i].in, lineNums)
		}
	}
}

func TestAssociationList(t *testing.T) {
	table := []struct {
		lines       []string
		names, vals []string
		errLine     int
	}{
		{[]string{"a=b"}, []string{"a"}, []string{"b"}, -1},
		{[]string{"a"}, []string{}, []string{}, 0},
		{[]string{"=b"}, []string{}, []string{}, 0},
		{[]string{"a=b", "c=", " a = "},
			[]string{"a", "c", "a"},
			[]string{"b", "", ""}, -1},
	}

	for i := range table {
		names, vals, errLine := associationList(table[i].lines)
		if errLine != table[i].errLine {
			t.Errorf("%d) Expected errLine = %d, got %d",
				i+1, table[i].errLine, errLine)
		}
		if errLine != -1 {
			continue
		}

		if !stringsEq(names, table[i].names) {
			t.Errorf("%d) Expected names = %v, got %v.",
				i+1, table[i].names, names)
		}
		if !stringsEq(vals, table[i].vals) {
			t.Errorf("%d) Expected vals = %v, got %v.",
				i+1, table[i].vals, vals)
		}

	}
}

func TestCheckDuplicateNames(t *testing.T) {
	table := []struct {
		names []string
		i, j  int
	}{
		{[]string{"a", "b", "c"}, -1, -1},
		{[]string{"a", "b", "b", "c", "c"}, 1, 2},
	}

	for k := range table {
		i, j := checkDuplicateNames(table[k].names)
		if i != table[k].i || j != table[k].j {
			t.Errorf("%d) expected (i, j) = (%d, %d) but got (%d, %d)",
				k+1, table[k].i, table[k].j, i, j)
		}
	}
}

func TestCheckValidNames(t *testing.T) {
	table := []struct {
		names, vars []string
		i           int
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c", "d"}, -1},
		{[]string{"a", "b", "c"}, []string{"a", "b", "d"}, 2},
		{[]string{"a", "a", "a"}, []string{"a", "b", "c", "d"}, -1},
	}

	for j := range table {
		vars := &ConfigVars{varNames: table[j].vars}
		i := checkValidNames(table[j].names, vars)
		if i != table[j].i {
			t.Errorf("%d) expected i = %d, but got %d", j+1, i, table[j].i)
		}
	}
}

func TestConvertAssoc(t *testing.T) {
	table := []struct {
		names, vals []string
		i           int
		xVal        int64
	}{
		{[]string{"a"}, []string{"3"}, -1, 3},
		{[]string{"a", "a"}, []string{"3", "meow"}, 1, 3},
	}

	config := struct{ x int64 }{}
	vars := NewConfigVars("meow")
	vars.Int(&config.x, "a", 0)

	for j := range table {
		config.x = 0
		i := convertAssoc(table[j].names, table[j].vals, vars)
		if i != table[j].i {
			t.Errorf("%d) expected errLine = %d, but got %d",
				j+1, table[j].i, i)
		}
		if i != -1 {
			continue
		}
		if config.x != table[j].xVal {
			t.Errorf("%d) expected config.x = %d, got %d",
				j+1, config.x, table[j].xVal)
		}
	}
}

func floatEq(x, y, eps float64) bool {
	return math.Abs(x-y) < eps
}

func floatsEq(xs, ys []float64, eps float64) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if !floatEq(xs[i], ys[i], eps) {
			return false
		}
	}
	return true
}

func boolsEq(xs, ys []bool) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if xs[i] != ys[i] {
			return false
		}
	}

	return true
}

func int64sEq(xs, ys []int64) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if xs[i] != ys[i] {
			return false
		}
	}
	return true
}

type testConfig struct {
	float  float64
	floats []float64
	num    int64
	nums   []int64
	okay   bool
	okays  []bool
	word   string
	words  []string
}

func makeTestConfig() (*testConfig, *ConfigVars) {
	config := &testConfig{}
	vars := NewConfigVars("config")
	vars.Int(&config.num, "num", 0)
	vars.Ints(&config.nums, "nums", []int64{})
	vars.Float(&config.float, "float", 0)
	vars.Floats(&config.floats, "floats", []float64{})
	vars.Bool(&config.okay, "okay", false)
	vars.Bools(&config.okays, "okays", []bool{})
	vars.String(&config.word, "word", "")
	vars.Strings(&config.words, "words", []string{})

	return config, vars
}

func TestValidConfig(t *testing.T) {
	config, vars := makeTestConfig()
	err := ReadConfig("config_test_files/success.config", vars)
	if err != nil {
		t.Errorf("Expected successful read of config file, but got "+
			"error:\n %s", err.Error())
	}

	if !floatEq(config.float, -1.2e4, 1) {
		t.Errorf("Expected float = %g, but got %g", -1.2e4, config.float)
	}
	if !floatsEq([]float64{2.5, 2.5, 2.5}, config.floats, 0.001) {
		t.Errorf("Expected floats = %v, but got %v.",
			[]float64{2.5, 2.5, 2.5}, config.floats)
	}

	if config.num != 3 {
		t.Errorf("Expected num = %d, but got %d", 3, config.num)
	}
	if !int64sEq(config.nums, []int64{1, 1, 2, 3, 5}) {
		t.Errorf("Expected nums = %v, but got %v",
			[]int64{1, 1, 2, 3, 5}, config.nums)
	}

	if config.okay != true {
		t.Errorf("Expected okay = %v, but got %v", true, config.okay)
	}
	if !boolsEq(config.okays, []bool{true, false, true}) {
		t.Errorf("Expected okays = %v, buf got %v",
			[]bool{true, false, true}, config.okays)
	}

	if config.word != "meow" {
		t.Errorf("Expected word = %v, but got %v", "meow", config.word)
	}
	if !stringsEq([]string{"dorothy", "maddy", "sahil"}, config.words) {
		t.Errorf("Expected words = %v, but got %v",
			[]string{"dorothy", "maddy", "sahil"}, config.words)
	}
}

func TestInvalidConfig(t *testing.T) {
	_, vars := makeTestConfig()

	fnames := []string{
		"config_test_files/empty.config",
		"config_test_files/wrong_header.config",
		"config_test_files/non_assignment.config",
		"config_test_files/no_variable.config",
		"config_test_files/dupicates.config",
		"config_test_files/invalid_var.config",
		"config_test_files/invalid_type.config",
	}

	for i := range fnames {
		err := ReadConfig(fnames[i], vars)
		if err == nil {
			t.Errorf("No error was reported when attempting to parse %s",
				fnames[i])
		} else if testing.Verbose() {
			fmt.Printf("%s:\n", fnames[i])
			fmt.Println(err.Error())
		}
	}
}

func TestValidFlags(t *testing.T) {
	config, vars := makeTestConfig()
	flags := []string{
		"--Num", "16",
		"--Nums", "1, 2, 3, 4, 5",
		"--Float", "16",
		"--Floats", "1", "2", "3", "4", "5",
		"--Okay", "true",
		"---Okays", "true, true", "false",
	}

	err := ReadFlags(flags, vars)
	if err != nil {
		t.Errorf("Could not parse valid flags: got the error '%s'", err.Error())
	}
	switch {
	case config.num != 16:
		t.Errorf("Flag Num not set.")
	case !int64sEq(config.nums, []int64{1, 2, 3, 4, 5}):
		t.Errorf("Flag Nums not set.")
	case config.float != 16:
		t.Errorf("Flag Float not set.")
	case !floatsEq(config.floats, []float64{1, 2, 3,4, 5}, 0.001):
		t.Errorf("Flag Floats not set.")
	case !config.okay:
		t.Errorf("Flag Okay not set.")
	case !boolsEq(config.okays, []bool{true, true, false}):
		t.Errorf("Flag Okay not set.")
	}
}
