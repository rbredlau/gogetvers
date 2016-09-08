package gogetvers

import (
	"path/filepath"
	"strings"
)

type Git struct {
	HomeDir   string
	ParentDir string
	Branch    string
	Hash      string
	OriginUrl string
	Describe  string
	Status    string
}

func (g *Git) StripDirPrefix(path string) {
	if g == nil {
		return
	}
	g.HomeDir = strings.TrimLeft(strings.Replace(g.HomeDir, path, "", -1), "\\/")
	g.ParentDir = strings.TrimLeft(strings.Replace(g.ParentDir, path, "", -1), "\\/")
}

func (g *Git) ToSlash() {
	if g == nil {
		return
	}
	g.HomeDir = filepath.ToSlash(g.HomeDir)
	g.ParentDir = filepath.ToSlash(g.ParentDir)
}

func (g *Git) FromSlash() {
	if g == nil {
		return
	}
	g.HomeDir = filepath.FromSlash(g.HomeDir)
	g.ParentDir = filepath.FromSlash(g.ParentDir)
}
