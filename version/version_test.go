package version

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		s                   string
		major, minor, patch int
		valid               bool
	}{
		{"0.0.0", 0, 0, 0, true},
		{"1.02.3", 1, 2, 3, true},
		{"", 0, 0, 0, false},
		{"0", 0, 0, 0, false},
		{"0.0", 0, 0, 0, false},
		{"0.0.0.0", 0, 0, 0, false},
		{"0.-1.0", 0, 0, 0, false},
	}

	for i := range tests {
		major, minor, patch, err := Parse(tests[i].s)
		if err != nil {
			if tests[i].valid {
				t.Errorf("Expected Parse('%s') to give an error, but it "+
					"doesn't.", tests[i].s)
			}
		} else {
			if !tests[i].valid {
				t.Errorf("Expected Parse('%s') to be valid, but it gave an "+
					"error.", tests[i].s)
			}
			if major != tests[i].major || minor != tests[i].minor ||
				patch != tests[i].patch {
				t.Errorf("Parse('%s') parsed to (%d, %d, %d).",
					tests[i].s, major, minor, patch)
			}
		}
	}
}

func TestLater(t *testing.T) {
	tests := []struct {
		s1, s2       string
		later, valid bool
	}{
		{"0.0.0", "0.0", false, false},
		{"0.0.0", "0.0.0", false, true},
		{"0.0.1", "0.0.0", true, true},
		{"0.1.0", "0.0.0", true, true},
		{"1.0.0", "0.0.0", true, true},
		{"0.0.0", "0.0.1", false, true},
		{"0.0.0", "0.1.0", false, true},
		{"0.0.0", "1.0.0", false, true},
		{"2.13.7", "2.12.19", true, true},
		{"2.12.19", "2.13.7", false, true},
	}

	for i := range tests {
		later, err := Later(tests[i].s1, tests[i].s2)
		if err == nil && !tests[i].valid {
			t.Errorf("Expected Later('%s', %s) to return an error, but it "+
				"didn't.", tests[i].s1, tests[i].s2)
		} else if err != nil && tests[i].valid {
			t.Errorf("Did not expect Later('%s', '%s') to return an error, "+
				"but it did.", tests[i].s1, tests[i].s2)
		} else if later != tests[i].later {
			t.Errorf("Later('%s', '%s') returned %v", tests[i].s1,
				tests[i].s2, tests[i].later)
		}
	}
}
