package gogetvers

import (
	"sort"
	"strings"
)

type GitList []*Git

func NewGitList(item ...*Git) GitList {
	return append([]*Git{}, item...)
}

func (gl GitList) Len() int {
	return len(gl)
}

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

func (gl GitList) Names() []string {
	rv := []string{}
	for _, v := range gl {
		rv = append(rv, v.HomeDir)
	}
	return rv
}

func (gl GitList) Sort() {
	sort.Sort(gl)
}

func (gl GitList) Swap(i, j int) {
	gl[i], gl[j] = gl[j], gl[i]
}
