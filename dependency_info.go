package gogetvers

import (
	"path/filepath"
	"strings"
)

type DependencyInfo struct {
	IsGo   bool   // True if not in GoSrcDir as a path.
	IsGit  bool   // True if .git info was found.
	Name   string // Name according to: go list -f {{.Deps}} from the parent package.
	DepDir string // Path to dependency.
	GitDir string // The .git directory.
}

func (d *DependencyInfo) StripDirPrefix(path string) {
	if d == nil {
		return
	}
	d.DepDir = strings.TrimLeft(strings.Replace(d.DepDir, path, "", -1), "\\/")
	d.GitDir = strings.TrimLeft(strings.Replace(d.GitDir, path, "", -1), "\\/")
}

func (d *DependencyInfo) ToSlash() {
	if d == nil {
		return
	}
	d.DepDir = filepath.ToSlash(d.DepDir)
	d.GitDir = filepath.ToSlash(d.GitDir)
}

func (d *DependencyInfo) FromSlash() {
	if d == nil {
		return
	}
	d.DepDir = filepath.FromSlash(d.DepDir)
	d.GitDir = filepath.FromSlash(d.GitDir)
}
