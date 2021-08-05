/*package version controls the version*/
package version

import (
	"fmt"
	"strconv"
	"strings"
)

// SourceVersion is the version string representing the semantic version number
// of the source code.
const SourceVersion = "1.0.3"

// Parse parses a semantic version number string and returns an error if
// the string is invalid.
func Parse(s string) (major, minor, patch int, err error) {
	toks := strings.Split(s, ".")
	errMsg := "Version string does not take the form of three " +
		"period-separated non-negative numbers"

	if len(toks) != 3 {
		return -1, -1, -1, fmt.Errorf(errMsg)
	}

	major, err = strconv.Atoi(toks[0])
	if err != nil {
		return -1, -1, -1, fmt.Errorf(errMsg)
	}
	minor, err = strconv.Atoi(toks[1])
	if err != nil {
		return -1, -1, -1, fmt.Errorf(errMsg)
	}
	patch, err = strconv.Atoi(toks[2])
	if err != nil {
		return -1, -1, -1, fmt.Errorf(errMsg)
	}

	if major < 0 || minor < 0 || patch < 0 {
		return -1, -1, -1, fmt.Errorf(errMsg)
	}

	return major, minor, patch, nil
}

// Greater returns true if s1 represents a later version of the source than
// s2. An error is returned if either is invalid.
func Later(s1, s2 string) (bool, error) {
	major1, minor1, patch1, err := Parse(s1)
	if err != nil {
		return false, err
	}
	major2, minor2, patch2, err := Parse(s2)
	if err != nil {
		return false, err
	}

	if major1 == major2 {
		if minor1 == minor2 {
			return patch1 > patch2, nil
		} else {
			return minor1 > minor2, nil
		}
	} else {
		return major1 > major2, nil
	}
}
