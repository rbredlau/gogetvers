package gogetvers

import (
	"path/filepath"
	"strings"
)

type pathsComposite struct {
	paths []*string
}

func newPathsComposite(paths ...*string) *pathsComposite {
	rv := &pathsComposite{paths: []*string{}}
	for _, v := range paths {
		rv.paths = append(rv.paths, v)
	}
	return rv
}

func (pc *pathsComposite) StripPathPrefix(prefix string) {
	if pc == nil {
		return
	}
	for _, v := range pc.paths {
		*v = strings.Replace(*v, prefix, "", 1)
	}
}

func (pc *pathsComposite) SetPathPrefix(prefix string) {
	if pc == nil {
		return
	}
	for _, v := range pc.paths {
		*v = filepath.Join(prefix, *v)
	}
}

func (pc *pathsComposite) PathsToSlash() {
	if pc == nil {
		return
	}
	for _, v := range pc.paths {
		*v = filepath.ToSlash(*v)
	}
}

func (pc *pathsComposite) PathsFromSlash() {
	if pc == nil {
		return
	}
	for _, v := range pc.paths {
		*v = filepath.FromSlash(*v)
	}
}
