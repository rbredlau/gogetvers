package gogetvers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PackageInfo summarizes a package and its dependencies.
type PackageInfo struct {
	PackageDir string // Package source directory; absolute path.
	RootDir    string // The root directory that contains everything.
	Git        *Git   // Git info for package.
	// Dependencies
	DepsBuiltin   []*BuiltinDependency
	DepsGit       []*GitDependency
	DepsUntracked []*UntrackedDependency
	//
	*PathsComposite
}

// Create a new PackageInfo by analyzing the package at packageDir
// with go-src path located at rootDir.
func NewPackageInfo(packageDir, rootDir string) *PackageInfo {
	rv := &PackageInfo{
		PackageDir:    packageDir,
		RootDir:       rootDir,
		DepsBuiltin:   []*BuiltinDependency{},
		DepsGit:       []*GitDependency{},
		DepsUntracked: []*UntrackedDependency{}}
	rv.SetPathsComposite()
	return rv
}

// Creates the embedded PathsComposite type within PackageInfo.
func (p *PackageInfo) SetPathsComposite() {
	if p != nil {
		p.PathsComposite = NewPathsComposite(&p.PackageDir, &p.RootDir)
		for _, dep := range p.DepsGit {
			dep.Git.SetPathsComposite()
		}
		p.Git.SetPathsComposite()
	}
}

// Opens the input file and decodes the manifest.
func LoadPackageInfoFile(inputFile string) (*PackageInfo, error) {
	if !IsFile(inputFile) {
		return nil, errors.New(fmt.Sprintf("Not a file @ %v", inputFile))
	}
	fr, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer fr.Close()
	//
	dec := json.NewDecoder(fr)
	summary := &PackageInfo{}
	err = dec.Decode(summary)
	if err != nil {
		return nil, err
	}
	// We have to reset paths composites as they aren't set in the JSON file.
	summary.SetPathsComposite()
	//
	return summary, nil
}

// Create a new package info type by analyzing a directory continaining the
// package.
func getPackageInfo(packageDir string, status *StatusWriter) (*PackageInfo, error) {
	// Absolute path.
	packageDir, err := filepath.Abs(packageDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	//
	status.Printf("Get package info for package @ %v\n", packageDir)
	// Get 'go list' information; this is package information according to golang.
	golist := NewCommandGoList()
	err = golist.Exec(packageDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	status.Printf("%v -> %v\n", golist.String(), golist.Output)
	// If we remove the output from golist from packageDir then
	// we'll have root directory of all sources.
	rootDir := strings.Replace(filepath.ToSlash(packageDir), golist.Output, "", -1)
	rootDir, err = filepath.Abs(rootDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	rootDir = strings.TrimRight(rootDir, "\\/")
	status.Printf("Root path @ %v\n", rootDir)
	// Get the git info for package.
	git, err := NewGitByFind(packageDir, rootDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	status.Writeln("Found package git information")
	// Get dependency information.
	golistdeps := NewCommandGoListDeps()
	err = golistdeps.Exec(packageDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	status.Printf("Dependencies are: %v\n", strings.Replace(golistdeps.Output, " ", ", ", -1))
	// Our return value.
	rv := NewPackageInfo(packageDir, rootDir)
	rv.Git = git
	// Get information for each dependency.
	status.Writeln("Getting dependency information...")
	status.Indent()
	deps := strings.Split(golistdeps.Output, " ")
	for _, depName := range deps {
		status.Printf("%v...", depName)
		dep, err := GetDependency(filepath.Join(rv.RootDir, depName), rv.RootDir)
		if err != nil {
			status.Error(err)
			return nil, err
		}
		switch d := dep.(type) {
		case *BuiltinDependency:
			status.Printf("built in\n")
			rv.DepsBuiltin = append(rv.DepsBuiltin, d)
		case *GitDependency:
			status.Printf("git\n")
			rv.DepsGit = append(rv.DepsGit, d)
		case *UntrackedDependency:
			status.Printf("untracked dependency\n")
			rv.DepsUntracked = append(rv.DepsUntracked, d)
		}
	}
	status.Outdent()
	status.Writeln("done")
	return rv, nil
}

// Return a package summary.
func (p *PackageInfo) getSummary() string {
	rv := "Package Summary\n"
	rv = rv + "    home> " + p.PackageDir + "\n"
	rv = rv + "    root> " + p.RootDir + "\n"
	rv = rv + "    gits>\n"
	if len(p.DepsGit) > 0 {
		rv = rv + "        " + strings.Join(p.getGitNames(), ", ") + "\n"
	}
	rv = rv + "    built ins>\n"
	if len(p.DepsBuiltin) > 0 {
		rv = rv + "        " + strings.Join(p.getBuiltinNames(), ", ") + "\n"
	}
	rv = rv + "    untracked>\n"
	if len(p.DepsUntracked) > 0 {
		rv = rv + "        " + strings.Join(p.getUntrackedNames(), ", ") + "\n"
	}
	if len(p.DepsGit) > 0 {
		rv = rv + "\n    git summary>\n"
		for _, git := range p.getGits() {
			rv = rv + "        " + strings.Replace(git.String(), "\n", "\n        ", -1) + "\n"
		}
	}
	return rv
}

// Condenses the gits in the package to a unique, sorted git list.
func (p *PackageInfo) getGits() GitList {
	gits := []*Git{}
	found := make(map[string]bool)
	if p == nil {
		return NewGitList(gits...)
	}
	// Dependency gits
	for _, dep := range p.DepsGit {
		if _, ok := found[dep.Git.HomeDir]; !ok {
			found[dep.Git.HomeDir] = true
			gits = append(gits, dep.Git)
		}
	}
	// Package git
	if _, ok := found[p.Git.HomeDir]; !ok {
		found[p.Git.HomeDir] = true
		gits = append(gits, p.Git)
	}
	// Return value sorted
	rv := NewGitList(gits...)
	rv.Sort()
	return rv
}

// Returns a list of gits on disk and gits missing from disk.
func (p *PackageInfo) getGitsDiskStatus() (exist GitList, dne GitList) {
	yeslist, nolist := []*Git{}, []*Git{}
	if p == nil {
		return NewGitList(yeslist...), NewGitList(nolist...)
	}
	//
	for _, v := range p.getGits() {
		if IsDir(v.HomeDir) {
			yeslist = append(yeslist, v)
		} else {
			nolist = append(nolist, v)
		}
	}
	// Return value sorted
	exist = NewGitList(yeslist...)
	exist.Sort()
	dne = NewGitList(nolist...)
	dne.Sort()
	return
}

// Returns three git lists: gits with local mods, gits without local mods, and gits
// not existing on disk.
func (p *PackageInfo) getGitsLocalModsStatus() (mods GitList, nomods GitList, dne GitList, rverr error) {
	yeslist, nolist, dnelist := []*Git{}, []*Git{}, []*Git{}
	if p == nil {
		return NewGitList(yeslist...), NewGitList(nolist...), NewGitList(dnelist...), errors.New("nil receiver")
	}
	//
	exist, dne := p.getGitsDiskStatus()
	for _, git := range exist {
		newgit, err := NewGit(git.HomeDir)
		if err != nil {
			return nil, nil, nil, err
		}
		if newgit.Status == "" {
			nolist = append(nolist, git)
		} else {
			yeslist = append(yeslist, git)
		}
	}
	// Return value sorted
	mods = NewGitList(yeslist...)
	mods.Sort()
	nomods = NewGitList(nolist...)
	nomods.Sort()
	dne.Sort()
	return
}

// Return a slice of git names.
func (p *PackageInfo) getGitNames() []string {
	if p == nil {
		return nil
	}
	rv := []string{}
	gits := p.getGits()
	for _, git := range gits {
		rv = append(rv, git.HomeDir)
	}
	sort.Strings(rv)
	return rv
}

// Return a slice of builtin names.
func (p *PackageInfo) getBuiltinNames() []string {
	if p == nil {
		return nil
	}
	rv := []string{}
	for _, bu := range p.DepsBuiltin {
		rv = append(rv, bu.Name)
	}
	sort.Strings(rv)
	return rv
}

// Return a slice of untracked names.
func (p *PackageInfo) getUntrackedNames() []string {
	if p == nil {
		return nil
	}
	rv := []string{}
	for _, un := range p.DepsUntracked {
		rv = append(rv, un.Name)
	}
	sort.Strings(rv)
	return rv
}

// Strip prefix from all path related members.
func (p *PackageInfo) StripPathPrefix(prefix string) {
	if p == nil {
		return
	}
	p.PathsComposite.StripPathPrefix(prefix)
	if p.Git != nil {
		p.Git.StripPathPrefix(prefix)
	}
	for _, dep := range p.DepsGit {
		dep.Git.StripPathPrefix(prefix)
	}
}

// Prepend prefix to all path related members.
func (p *PackageInfo) SetPathPrefix(prefix string) {
	if p == nil {
		return
	}
	p.PathsComposite.SetPathPrefix(prefix)
	if p.Git != nil {
		p.Git.SetPathPrefix(prefix)
	}
	for _, dep := range p.DepsGit {
		dep.Git.SetPathPrefix(prefix)
	}
}

// Forward PathsToSlash() to all appropriate members.
func (p *PackageInfo) PathsToSlash() {
	if p == nil {
		return
	}
	p.PathsComposite.PathsToSlash()
	if p.Git != nil {
		p.Git.PathsToSlash()
	}
	for _, dep := range p.DepsGit {
		dep.Git.PathsToSlash()
	}
}

// Forward PathsFromSlash() to all appropriate members.
func (p *PackageInfo) PathsFromSlash() {
	if p == nil {
		return
	}
	p.PathsComposite.PathsFromSlash()
	if p.Git != nil {
		p.Git.PathsFromSlash()
	}
	for _, dep := range p.DepsGit {
		dep.Git.PathsFromSlash()
	}
}
