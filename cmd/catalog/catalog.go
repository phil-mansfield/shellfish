package catalog

import (
	"fmt"
	"strconv"
	"strings"
)

func CommentString(intNames, floatNames []string, order []int) string {
	tokens := []string{"# Column contents:"}
	n := 0
	for i := range intNames {
		tokens = append(tokens, fmt.Sprintf("%s(%d)", intNames[i], n))
		n++
	}
	for i := range floatNames {
		tokens = append(tokens, fmt.Sprintf("%s(%d)", floatNames[i], n))
		n++
	}

	return strings.Join(tokens, " ")
}

func FormatCols(intCols [][]int, floatCols [][]float64, order []int) []string {
	if (len(intCols) == 0 && len(floatCols) == 0) ||
		(len(intCols) > 0 && len(intCols[0]) == 0) ||
		(len(floatCols) > 0 && len(floatCols[0]) == 0){
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
		if idx >= len(intCols) + len(floatCols) {
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
	tokens := make([]string, len(intCols) + len(floatCols))
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
		if n > width { width = n }
	}

	out := []string{}
	for i := range col {
		out = append(out, fmt.Sprintf("%*d", width, col[i]))
	}

	return out
}

func formatFloatCol(col []float64) []string {
	width := len(fmt.Sprintf("%.4g", col[0]))
	for i := 1; i < len(col); i++ {
		n := len(fmt.Sprintf("%.4g", col[i]))
		if n > width { width = n }
	}

	out := []string{}
	for i := range col {
		out = append(out, fmt.Sprintf("%*.4g", width, col[i]))
	}

	return out
}

func Uncomment(lines []string) (out []string, lineNums []int) {
	for i := range lines {
		idx := strings.Index(lines[i], "#")
		if idx >= 0 {
			lines[i] = lines[i][:idx]
		}
	}

	out = []string{}
	lineNums = []int{}
	for i := range lines {
		trimmed := strings.Trim(lines[i], " \t")
		if len(trimmed) > 0 {
			out = append(out, trimmed)
			lineNums = append(lineNums, i + 1)
		}
	}
	return out, lineNums
}

func ParseCols(
	lines []string, intIdxs, floatIdxs []int,
) ([][]int, [][]float64, error) {
	if len(intIdxs) == 0 && len(floatIdxs) == 0 { return nil, nil, nil }

	fLines, lineNums := Uncomment(lines)
	minWidth := -1
	for _, x := range intIdxs {
		if x > minWidth { minWidth = x }
	}
	for _, x := range floatIdxs {
		if x > minWidth { minWidth = x }
	}
	minWidth++

	intCols := make([][]int, len(intIdxs))
	floatCols := make([][]float64, len(floatIdxs))
	
	for i := range fLines {
		toks := tokenize(fLines[i])

		if len(toks) < minWidth {
			return nil, nil, fmt.Errorf(
				"Line %d has %d columns, but I need %d columns.",
				lineNums[i], len(toks), minWidth,
			)
		} else {
			for colIdx, j := range intIdxs {
				n, err := strconv.Atoi(toks[j])
				if err != nil {
					return nil, nil, fmt.Errorf("Cannot parse column %d of " +
						"line %d, '%s', to an int.", j, lineNums[i], toks[j])
				}
				intCols[colIdx] = append(intCols[j], n)
			}

			for colIdx, j := range floatIdxs {
				x, err := strconv.ParseFloat(toks[j], 64)
				if err != nil {
					return nil, nil, fmt.Errorf("Cannot parse column %d of " +
						"line %d, '%s', to a float.", j, lineNums[i], toks[j])
				}
				floatCols[colIdx] = append(floatCols[colIdx], x)
			}
		}
	}

	return intCols, floatCols, nil
}

func tokenize(line string) []string {
	toks := strings.Split(line, " ")
	fToks := []string{}
	for i := range toks {
		if len(toks[i]) > 0 {
			fToks = append(fToks, toks[i])
		}
	}
	return fToks
}
