package gogetvers

import (
	"path/filepath"
	"strings"
)

// Other types embed this type to avoid cut and paste.
type PathsComposite struct {
	Paths []*string
}

// Creates a new PathsComposite type.
func NewPathsComposite(paths ...*string) *PathsComposite {
	rv := &PathsComposite{Paths: []*string{}}
	for _, v := range paths {
		rv.Paths = append(rv.Paths, v)
	}
	return rv
}

// Removes prefix from all of the paths within PathsComposite.
func (pc *PathsComposite) StripPathPrefix(prefix string) {
	if pc == nil {
		return
	}
	for _, v := range pc.Paths {
		*v = strings.Replace(*v, prefix, "", 1)
		*v = strings.TrimLeft(*v, "\\/")
	}
}

// Preprends prefix to all paths within PathsComposite.
func (pc *PathsComposite) SetPathPrefix(prefix string) {
	if pc == nil {
		return
	}
	for _, v := range pc.Paths {
		*v = filepath.Join(prefix, *v)
	}
}

// Calls filepath.ToSlash() on all paths within PathsComposite.
func (pc *PathsComposite) PathsToSlash() {
	if pc == nil {
		return
	}
	for _, v := range pc.Paths {
		*v = filepath.ToSlash(*v)
	}
}

// Calls filepath.FromSlash() on all paths within PathsComposite.
func (pc *PathsComposite) PathsFromSlash() {
	if pc == nil {
		return
	}
	for _, v := range pc.Paths {
		*v = filepath.FromSlash(*v)
	}
}
