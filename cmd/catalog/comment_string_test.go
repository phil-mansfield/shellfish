package catalog

import (
	"testing"
)

func TestCommentString(t *testing.T) {
	tests := []struct {
		intNames, floatNames []string
		order []int
		sizes []int
		out string
	} {
		{[]string{"A"}, []string{}, []int{0}, []int{1},
			"# Column contents: A(0)"},
		{[]string{}, []string{"A"}, []int{0}, []int{1},
			"# Column contents: A(0)"},
		{[]string{"A"}, []string{}, []int{0}, []int{11},
			"# Column contents: A(0-10)"},
		{[]string{"A"}, []string{"B"}, []int{0, 1}, []int{1, 1},
			"# Column contents: A(0) B(1)"},
		{[]string{"A"}, []string{"B"}, []int{1, 0}, []int{1, 1},
			"# Column contents: B(0) A(1)"},
		{[]string{"A"}, []string{"B"}, []int{0, 1}, []int{1, 2},
			"# Column contents: A(0) B(1-2)"},
		{[]string{"A"}, []string{"B"}, []int{1, 0}, []int{1, 2},
			"# Column contents: B(0-1) A(2)"},
		{[]string{"A", "C"}, []string{"B"}, []int{0, 1}, []int{1, 2},
			"# Column contents: A(0) B(1-2) C(3)"},
	}

	for i, test := range tests {
		out := CommentString(test.intNames,
			test.floatNames, test.order, test.sizes)
		if out != test.out {
			t.Errorf("%d) Expected '%s', got '%s'.", i, test.out, out)
		}
	}
}