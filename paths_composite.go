package gogetvers

import (
	"path/filepath"
	"strings"
)

type PathsComposite struct {
	Paths []*string
}

func NewPathsComposite(paths ...*string) *PathsComposite {
	rv := &PathsComposite{Paths: []*string{}}
	for _, v := range paths {
		rv.Paths = append(rv.Paths, v)
	}
	return rv
}

func (pc *PathsComposite) StripPathPrefix(prefix string) {
	if pc == nil {
		return
	}
	for _, v := range pc.Paths {
		*v = strings.Replace(*v, prefix, "", 1)
		*v = strings.TrimLeft(*v, "\\/")
	}
}

func (pc *PathsComposite) SetPathPrefix(prefix string) {
	if pc == nil {
		return
	}
	for _, v := range pc.Paths {
		*v = filepath.Join(prefix, *v)
	}
}

func (pc *PathsComposite) PathsToSlash() {
	if pc == nil {
		return
	}
	for _, v := range pc.Paths {
		*v = filepath.ToSlash(*v)
	}
}

func (pc *PathsComposite) PathsFromSlash() {
	if pc == nil {
		return
	}
	for _, v := range pc.Paths {
		*v = filepath.FromSlash(*v)
	}
}
