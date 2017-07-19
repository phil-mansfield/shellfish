package catalog

import (
	"fmt"
	"os"
	"io/ioutil"
	"strconv"
	"bytes"
	"strings"
)

func CommentString(
	intNames, floatNames []string, order, sizes []int,
) string {

	tokens := []string{"# Column contents:"}
	for i := range intNames {
		tokens = append(tokens, fmt.Sprintf("%s", intNames[i]))
	}
	for i := range floatNames {
		tokens = append(tokens, fmt.Sprintf("%s", floatNames[i]))
	}

	orderedTokens := []string{tokens[0]}
	orderedSizes := []int{}
	for _, idx := range order {
		if idx >= len(intNames)+len(floatNames) {
			panic("Column ordering out of range.")
		}

		orderedTokens = append(orderedTokens, tokens[idx+1])
		orderedSizes = append(orderedSizes, sizes[idx])

	}

	n := 0
	for i := 1; i < len(orderedTokens); i++ {
		if orderedSizes[i-1] == 1 {
			orderedTokens[i] = fmt.Sprintf("%s(%d)", orderedTokens[i], n)
		} else {
			orderedTokens[i] = fmt.Sprintf("%s(%d-%d)", orderedTokens[i],
				n, n+orderedSizes[i-1]-1)
		}
		n += orderedSizes[i-1]
	}

	return strings.Join(orderedTokens, " ")
}

func FormatCols(intCols [][]int, floatCols [][]float64, order []int) []string {
	if (len(intCols) == 0 && len(floatCols) == 0) ||
		(len(intCols) > 0 && len(intCols[0]) == 0) ||
		(len(floatCols) > 0 && len(floatCols[0]) == 0) {
		return []string{}
	}

	formattedIntCols := make([][]string, len(intCols))
	formattedFloatCols := make([][]string, len(floatCols))

	height := -1
	for i := range intCols {
		formattedIntCols[i] = formatIntCol(intCols[i])
		if height == -1 {
			height = len(intCols[i])
		} else if height != len(intCols[i]) {
			panic("Columns of unequal height.")
		}
	}

	for i := range floatCols {
		formattedFloatCols[i] = formatFloatCol(floatCols[i])
		if height == -1 {
			height = len(floatCols[i])
		} else if height != len(floatCols[i]) {
			panic("Columns of unequal height.")
		}
	}

	orderedCols := [][]string{}
	for _, idx := range order {
		if idx >= len(intCols)+len(floatCols) {
			panic("Column ordering out of range.")
		}

		if idx < len(intCols) {
			orderedCols = append(orderedCols, formattedIntCols[idx])
		} else {
			idx -= len(intCols)
			orderedCols = append(orderedCols, formattedFloatCols[idx])
		}
	}

	lines := []string{}
	tokens := make([]string, len(intCols)+len(floatCols))
	for i := 0; i < height; i++ {
		for j := range orderedCols {
			tokens[j] = orderedCols[j][i]
		}
		line := strings.Join(tokens, " ")
		lines = append(lines, line)
	}

	return lines
}

func formatIntCol(col []int) []string {
	width := len(fmt.Sprintf("%d", col[0]))
	for i := 1; i < len(col); i++ {
		n := len(fmt.Sprintf("%d", col[i]))
		if n > width {
			width = n
		}
	}

	out := []string{}
	for i := range col {
		out = append(out, fmt.Sprintf("%*d", width, col[i]))
	}

	return out
}

func formatFloatCol(col []float64) []string {
	width := len(fmt.Sprintf("%.6g", col[0]))
	for i := 1; i < len(col); i++ {
		n := len(fmt.Sprintf("%.6g", col[i]))
		if n > width {
			width = n
		}
	}

	out := []string{}
	for i := range col {
		out = append(out, fmt.Sprintf("%*.6g", width, col[i]))
	}

	return out
}

// Parse parses the specified columns in a byte block.
func Parse(data []byte, icolIdxs, fcolIdxs []int) (
[][]int, [][]float64, error,
) {
	lines, nComm := split(data, '\n', '#')
	lines = uncomment(lines, '#', nComm)
	lines = trim(lines, ' ')
	return parse(lines, ' ', icolIdxs, fcolIdxs)
}

func ReadFile(fname string, icolIdxs, fcolIdxs []int) (
[][]int, [][]float64, error,
) {
	f, err := os.Open(fname)
	if err != nil { return nil, nil, err }
	data, err := ioutil.ReadAll(f)
	if err != nil { return nil, nil, err }
	return Parse(data, icolIdxs, fcolIdxs)
}

func ReadStdin(fname string, icolIdxs, fcolIdxs []int) (
[][]int, [][]float64, error,
) {
	data, err  := ioutil.ReadAll(os.Stdin)
	if err != nil { return nil, nil, err }
	return Parse(data, icolIdxs, fcolIdxs)
}

// split splits a byte splice at each separating flag. Faster than
// bytes.Split() because slicing is used instead of allocations and because
// only one separator is used.
//
// Some of the calculations associated with uncommenting are done here for a
// slight performance boost.
func split(data []byte, sep, comm byte) (lines [][]byte, nComm int) {
	n, nComm := 0, 0
	for _, c := range data {
		if c == sep { n++ }
		if c == comm { nComm++ }
	}

	tokens := make([][]byte, n+1)

	idx := 0
	for j := 0; j < n; j++ {
		data = data[idx:]
		idx = bytes.IndexByte(data, sep)
		tokens[j] = data[:idx]
		idx++
	}
	tokens[n] = data[idx:]

	return tokens, nComm
}

// uncomment removes file comments  in the form of "data # comment". Optimized
// for the common case where comments are rare and at the start of the file.
func uncomment(lines [][]byte, comm byte, nComm int) [][]byte {
	if nComm == 0 { return lines }

	for i, line := range lines {
		commentStart := bytes.IndexByte(line, comm)
		if commentStart == -1 {
			continue
		}

		lines[i] = line[:commentStart]

		n := 1
		for _, c := range line[commentStart+1:] {
			if c == comm { n++ }
		}

		nComm -= n
		if nComm == 0 { return lines }
	}

	return lines
}

// trim removes empty lines.
func trim(lines [][]byte, sep byte) [][]byte {
	j := 0

	LineLoop:
	for i, line := range lines {
		for _, c := range line {
			if c != sep {
				lines[j] = lines[i]
				j++
				continue LineLoop
			}
		}
	}

	return lines[:j]
}

func parse(lines [][]byte, sep byte, icolIdxs, fcolIdxs []int) (
[][]int, [][]float64, error,
) {
	// Set up output and buffers

	icols := make([][]int, len(icolIdxs))
	fcols := make([][]float64, len(fcolIdxs))

	for i := range icols { icols[i] = make([]int, len(lines)) }
	for i := range fcols { fcols[i] = make([]float64, len(lines)) }

	if len(lines) == 0 { return icols, fcols, nil }
	buf := make([][]byte, len(bytes.Fields(lines[0])))


	var err error
	for i, line := range lines {

		// Break line up into fields/words

		words := fields(line, sep, buf)
		if len(words) != len(buf) {
			return nil, nil, fmt.Errorf(
				"Data (not file) line %d has %d columns, not %d.",
				i+1, len(words), len(buf),
			)
		}

		// Parse strings.

		for j := range icolIdxs {
			icols[j][i], err = strconv.Atoi(
				string(words[icolIdxs[j]]),
			)
			if err != nil { return nil, nil, err }
		}
		for j := range fcolIdxs {
			fcols[j][i], err = strconv.ParseFloat(
				string(words[fcolIdxs[j]]), 64,
			)
			if err != nil { return nil, nil, err }
		}
	}

	return icols, fcols, nil
}

// Optimized and buffered analog to the standard library's bytes.FieldsFunc()
// function.
func fields(data []byte, sep byte, buf [][]byte) [][]byte {
	n := 0
	inField := false
	for _, c := range data {
		wasInField := inField
		inField = sep != c
		if inField && !wasInField { n++ }
	}

	na := 0
	fieldStart := -1

	for i := 0; i < len(data) && na < n; i++ {
		c := data[i]

		if fieldStart < 0 && c != sep {
			fieldStart = i
			continue
		}

		if fieldStart >= 0 && c == sep {
			buf[na] = data[fieldStart: i]
			na++
			fieldStart = -1
		}
	}

	if fieldStart >= 0 {
		buf[na] = data[fieldStart: len(data)]
		na++
	}

	return buf[0:na]
}
