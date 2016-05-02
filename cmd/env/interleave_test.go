package env

import (
	"testing"
)

type interleaveTest struct {
	cols   [][]interface{}
	isSnap []bool
	out    [][][]interface{}
}

func TestEmptyInterleave(t *testing.T) {
	tests := []interleaveTest{
		{
			cols:   [][]interface{}{[]interface{}{}},
			isSnap: []bool{true},
			out:    [][][]interface{}{},
		},
		{
			cols:   [][]interface{}{[]interface{}{}, []interface{}{}},
			isSnap: []bool{true, false},
			out:    [][][]interface{}{},
		},
	}

	for i := range tests {
		cols, isSnap, out := tests[i].cols, tests[i].isSnap, tests[i].out
		res := interleave(cols, isSnap)
		if !interleaveOutputEq(res, out) {
			t.Errorf("%d) Expected %v, but got %v.", i, out, res)
		}
	}
}

func TestSingletonInterleave(t *testing.T) {
	tests := []interleaveTest{
		{
			cols:   [][]interface{}{[]interface{}{1, 2, 3}},
			isSnap: []bool{true},
			out: [][][]interface{}{
				[][]interface{}{[]interface{}{1}},
				[][]interface{}{[]interface{}{2}},
				[][]interface{}{[]interface{}{3}},
			},
		},
	}

	for i := range tests {
		cols, isSnap, out := tests[i].cols, tests[i].isSnap, tests[i].out
		res := interleave(cols, isSnap)
		if !interleaveOutputEq(res, out) {
			t.Errorf("%d) Expected %v, but got %v.", i, out, res)
		}
	}
}

func TestPairInterleave(t *testing.T) {
	tests := []interleaveTest{
		{
			cols: [][]interface{}{[]interface{}{1, 2, 3},
				[]interface{}{4, 5, 6}},
			isSnap: []bool{true, true},
			out: [][][]interface{}{
				[][]interface{}{[]interface{}{1, 4}},
				[][]interface{}{[]interface{}{2, 5}},
				[][]interface{}{[]interface{}{3, 6}},
			},
		},
		{
			cols: [][]interface{}{[]interface{}{1, 2, 3},
				[]interface{}{4}},
			isSnap: []bool{true, false},
			out: [][][]interface{}{
				[][]interface{}{[]interface{}{1, 4}},
				[][]interface{}{[]interface{}{2, 4}},
				[][]interface{}{[]interface{}{3, 4}},
			},
		},
		{
			cols: [][]interface{}{[]interface{}{1, 2, 3},
				[]interface{}{4, 5}},
			isSnap: []bool{true, false},
			out: [][][]interface{}{
				[][]interface{}{[]interface{}{1, 4}, []interface{}{1, 5}},
				[][]interface{}{[]interface{}{2, 4}, []interface{}{2, 5}},
				[][]interface{}{[]interface{}{3, 4}, []interface{}{3, 5}},
			},
		},
		{
			cols:   [][]interface{}{[]interface{}{4, 5}, []interface{}{1, 2, 3}},
			isSnap: []bool{false, true},
			out: [][][]interface{}{
				[][]interface{}{[]interface{}{4, 1}, []interface{}{5, 1}},
				[][]interface{}{[]interface{}{4, 2}, []interface{}{5, 2}},
				[][]interface{}{[]interface{}{4, 3}, []interface{}{5, 3}},
			},
		},
	}

	for i := range tests {
		cols, isSnap, out := tests[i].cols, tests[i].isSnap, tests[i].out
		res := interleave(cols, isSnap)
		if !interleaveOutputEq(res, out) {
			t.Errorf("%d) Expected %v, but got %v.", i, out, res)
		}
	}
}

func TestMultiInterleave(t *testing.T) {
	tests := []interleaveTest{
		{
			cols: [][]interface{}{[]interface{}{1, 2},
				[]interface{}{3, 4}, []interface{}{5, 6}},
			isSnap: []bool{true, false, false},
			out: [][][]interface{}{
				[][]interface{}{[]interface{}{1, 3, 5}, []interface{}{1, 4, 5},
					[]interface{}{1, 3, 6}, []interface{}{1, 4, 6}},
				[][]interface{}{[]interface{}{2, 3, 5}, []interface{}{2, 4, 5},
					[]interface{}{2, 3, 6}, []interface{}{2, 4, 6}},
			},
		},
		{
			cols: [][]interface{}{[]interface{}{1, 2},
				[]interface{}{3, 4}, []interface{}{5}, []interface{}{6, 7, 8}},
			isSnap: []bool{true, false, false, false},
			out: [][][]interface{}{
				[][]interface{}{
					[]interface{}{1, 3, 5, 6},
					[]interface{}{1, 4, 5, 6},
					[]interface{}{1, 3, 5, 7},
					[]interface{}{1, 4, 5, 7},
					[]interface{}{1, 3, 5, 8},
					[]interface{}{1, 4, 5, 8},
				},
				[][]interface{}{
					[]interface{}{2, 3, 5, 6},
					[]interface{}{2, 4, 5, 6},
					[]interface{}{2, 3, 5, 7},
					[]interface{}{2, 4, 5, 7},
					[]interface{}{2, 3, 5, 8},
					[]interface{}{2, 4, 5, 8},
				},
			},
		},
	}

	for i := range tests {
		cols, isSnap, out := tests[i].cols, tests[i].isSnap, tests[i].out
		res := interleave(cols, isSnap)
		if !interleaveOutputEq(res, out) {
			t.Errorf("%d) Expected %v, but got %v.", i, out, res)
		}
	}
}

func interleaveOutputEq(xs, ys [][][]interface{}) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if len(xs[i]) != len(ys[i]) {
			return false
		}
		for j := range xs[i] {
			if !intInterfacesEq(xs[i][j], ys[i][j]) {
				return false
			}
		}
	}
	return true
}

func intInterfacesEq(xs, ys []interface{}) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		x, ok := xs[i].(int)
		if !ok {
			return false
		}
		y, ok := ys[i].(int)
		if !ok {
			return false
		}
		if x != y {
			return false
		}
	}
	return true
}
