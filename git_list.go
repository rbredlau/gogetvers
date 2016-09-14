package gogetvers

import (
	"sort"
	"strings"
)

// A utility type to sort Git types.
type GitList []*Git

// Create a new GitList
func NewGitList(item ...*Git) GitList {
	return append([]*Git{}, item...)
}

// Satisfies Sort interface.
func (gl GitList) Len() int {
	return len(gl)
}

// Satisfies Sort interface; one Git type is considered less
// than another if it is higher up in the file system or
// alphabetically if they have the same parent folder.
func (gl GitList) Less(i, j int) bool {
	gl[i].PathsToSlash()
	gl[j].PathsToSlash()
	a, b := gl[i].HomeDir, strings.Count(gl[i].HomeDir, "/")
	z, y := gl[j].HomeDir, strings.Count(gl[j].HomeDir, "/")
	if b == y {
		// Same number of delims, sort by value
		if strings.Compare(a, z) < 0 {
			return true
		} else {
			return false
		}
	}
	return b < y
}

// Returns a list of all HomeDir for the gits in the list.
func (gl GitList) Names() []string {
	rv := []string{}
	for _, v := range gl {
		rv = append(rv, v.HomeDir)
	}
	return rv
}

// Sorts the list.
func (gl GitList) Sort() {
	sort.Sort(gl)
}

// Satisfies Sort interface.
func (gl GitList) Swap(i, j int) {
	gl[i], gl[j] = gl[j], gl[i]
}
