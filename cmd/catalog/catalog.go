package catalog

import (
	"fmt"
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
	for i := range intCols {
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
		if idx >= len(intCols) {
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
		out = append(out, fmt.Sprintf("%*d", col[i], width))
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
		out = append(out, fmt.Sprintf("%*.4g", col[i], width))
	}

	return out
}

func Uncomment(lines []string) []string {
	for i := range lines {
		idx := strings.Index(lines[i], "#")
		if idx >= 0 {
			lines[i] = lines[i][:idx]
		}
	}

	out := []string{}
	for i := range lines {
		trimmed := strings.Trim(lines[i], " \t")
		if len(trimmed) > 0 {
			out = append(out, trimmed)
		}
	}
	return out
}
