package gogetvers

import (
	"path/filepath"
	"strings"
)

type PackageInfo struct {
	PackageDir  string                     // Package source directory; absolute path.
	GitDir      string                     // Path to .git for package.
	Git         *Git                       // Git info for package.
	Deps        []string                   // List of package dependencies.
	GoSrcDir    string                     // Absolute path of Go src that contains SourceDir.
	DepInfo     map[string]*DependencyInfo // Map of dependency info.
	GitDirs     map[string][]*DependencyInfo
	Gits        map[string]*Git
	Untrackable map[string]*DependencyInfo
}

func (p *PackageInfo) StripDirPrefix() {
	if p == nil {
		return
	}
	p.PackageDir = strings.TrimLeft(strings.Replace(p.PackageDir, p.GoSrcDir, "", -1), "\\/")
	p.GitDir = strings.TrimLeft(strings.Replace(p.GitDir, p.GoSrcDir, "", -1), "\\/")
	if p.Git != nil {
		p.Git.StripDirPrefix(p.GoSrcDir)
	}
	for _, v := range p.DepInfo {
		v.StripDirPrefix(p.GoSrcDir)
	}
	for _, v := range p.Gits {
		v.StripDirPrefix(p.GoSrcDir)
	}
}

func (p *PackageInfo) ToSlash() {
	if p == nil {
		return
	}
	p.PackageDir = filepath.ToSlash(p.PackageDir)
	p.GitDir = filepath.ToSlash(p.GitDir)
	p.GoSrcDir = filepath.ToSlash(p.GoSrcDir)
	if p.Git != nil {
		p.Git.ToSlash()
	}
	for _, v := range p.DepInfo {
		v.ToSlash()
	}
	for _, v := range p.Gits {
		v.ToSlash()
	}
}

func (p *PackageInfo) FromSlash() {
	if p == nil {
		return
	}
	p.PackageDir = filepath.FromSlash(p.PackageDir)
	p.GitDir = filepath.FromSlash(p.GitDir)
	p.GoSrcDir = filepath.FromSlash(p.GoSrcDir)
	if p.Git != nil {
		p.Git.FromSlash()
	}
	for _, v := range p.DepInfo {
		v.FromSlash()
	}
	for _, v := range p.Gits {
		v.FromSlash()
	}
}
