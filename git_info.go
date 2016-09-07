package gogetvers

import (
	"path/filepath"
	"strings"
)

type GitInfo struct {
	HomeDir   string
	ParentDir string
	Branch    string
	Hash      string
	OriginUrl string
	Describe  string
	Status    string
}

func (g *GitInfo) StripGoSrcDir(path string) {
	if g == nil {
		return
	}
	g.HomeDir = strings.TrimLeft(strings.Replace(g.HomeDir, path, "", -1), "\\/")
	g.ParentDir = strings.TrimLeft(strings.Replace(g.ParentDir, path, "", -1), "\\/")
}

func (g *GitInfo) ToSlash() {
	if g == nil {
		return
	}
	g.HomeDir = filepath.ToSlash(g.HomeDir)
	g.ParentDir = filepath.ToSlash(g.ParentDir)
}

func (g *GitInfo) FromSlash() {
	if g == nil {
		return
	}
	g.HomeDir = filepath.FromSlash(g.HomeDir)
	g.ParentDir = filepath.FromSlash(g.ParentDir)
}
