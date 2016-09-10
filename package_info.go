package gogetvers

import (
	fs "broadlux/fileSystem"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type PackageInfo struct {
	PackageDir string // Package source directory; absolute path.
	RootDir    string // The root directory that contains everything.
	Git        *Git   // Git info for package.
	// Dependencies
	DepsBuiltin   []*builtinDependency
	DepsGit       []*gitDependency
	DepsUntracked []*untrackedDependency
	//
	*pathsComposite
}

func newPackageInfo(packageDir, rootDir string) *PackageInfo {
	rv := &PackageInfo{
		PackageDir:    packageDir,
		RootDir:       rootDir,
		DepsBuiltin:   []*builtinDependency{},
		DepsGit:       []*gitDependency{},
		DepsUntracked: []*untrackedDependency{}}
	rv.setPathsComposite()
	return rv
}

func (p *PackageInfo) setPathsComposite() {
	if p != nil {
		p.pathsComposite = newPathsComposite(&p.PackageDir, &p.RootDir)
		for _, dep := range p.DepsGit {
			dep.Git.setPathsComposite()
		}
		p.Git.setPathsComposite()
	}
}

// Opens the input file and decodes the manifest.
func loadPackageInfoFile(inputFile string) (*PackageInfo, error) {
	if !fs.IsFile(inputFile) {
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
	summary.setPathsComposite()
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
	golist := newCommandGoList()
	err = golist.exec(packageDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	status.Printf("%v -> %v\n", golist.String(), golist.output)
	// If we remove the output from golist from packageDir then
	// we'll have root directory of all sources.
	rootDir := strings.Replace(filepath.ToSlash(packageDir), golist.output, "", -1)
	rootDir, err = filepath.Abs(rootDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	rootDir = strings.TrimRight(rootDir, "\\/")
	status.Printf("Root path @ %v\n", rootDir)
	// Get the git info for package.
	git, err := newGitByFind(packageDir, rootDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	status.Writeln("Found package git information")
	// Get dependency information.
	golistdeps := newCommandGoListDeps()
	err = golistdeps.exec(packageDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	status.Printf("Dependencies are: %v\n", strings.Replace(golistdeps.output, " ", ", ", -1))
	// Our return value.
	rv := newPackageInfo(packageDir, rootDir)
	rv.Git = git
	// Get information for each dependency.
	status.Writeln("Getting dependency information...")
	status.Indent()
	deps := strings.Split(golistdeps.output, " ")
	for _, depName := range deps {
		status.Printf("%v...", depName)
		dep, err := getDependency(filepath.Join(rv.RootDir, depName), rv.RootDir)
		if err != nil {
			status.Error(err)
			return nil, err
		}
		switch d := dep.(type) {
		case *builtinDependency:
			status.Printf("built in\n")
			rv.DepsBuiltin = append(rv.DepsBuiltin, d)
		case *gitDependency:
			status.Printf("git\n")
			rv.DepsGit = append(rv.DepsGit, d)
		case *untrackedDependency:
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
		return newGitList(gits...)
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
	rv := newGitList(gits...)
	rv.Sort()
	return rv
}

// Returns a list of gits on disk and gits missing from disk.
func (p *PackageInfo) getGitsDiskStatus() (exist GitList, dne GitList) {
	yeslist, nolist := []*Git{}, []*Git{}
	if p == nil {
		return newGitList(yeslist...), newGitList(nolist...)
	}
	//
	for _, v := range p.getGits() {
		if fs.IsDir(v.HomeDir) {
			yeslist = append(yeslist, v)
		} else {
			nolist = append(nolist, v)
		}
	}
	// Return value sorted
	exist = newGitList(yeslist...)
	exist.Sort()
	dne = newGitList(nolist...)
	dne.Sort()
	return
}

// Returns three git lists: gits with local mods, gits without local mods, and gits
// not existing on disk.
func (p *PackageInfo) getGitsLocalModsStatus() (mods GitList, nomods GitList, dne GitList, rverr error) {
	yeslist, nolist, dnelist := []*Git{}, []*Git{}, []*Git{}
	if p == nil {
		return newGitList(yeslist...), newGitList(nolist...), newGitList(dnelist...), errors.New("nil receiver")
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
	mods = newGitList(yeslist...)
	mods.Sort()
	nomods = newGitList(nolist...)
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

func (p *PackageInfo) StripPathPrefix(prefix string) {
	if p == nil {
		return
	}
	p.pathsComposite.StripPathPrefix(prefix)
	if p.Git != nil {
		p.Git.StripPathPrefix(prefix)
	}
	for _, dep := range p.DepsGit {
		dep.Git.StripPathPrefix(prefix)
	}
}

func (p *PackageInfo) SetPathPrefix(prefix string) {
	if p == nil {
		return
	}
	p.pathsComposite.SetPathPrefix(prefix)
	if p.Git != nil {
		p.Git.SetPathPrefix(prefix)
	}
	for _, dep := range p.DepsGit {
		dep.Git.SetPathPrefix(prefix)
	}
}

func (p *PackageInfo) PathsToSlash() {
	if p == nil {
		return
	}
	p.pathsComposite.PathsToSlash()
	if p.Git != nil {
		p.Git.PathsToSlash()
	}
	for _, dep := range p.DepsGit {
		dep.Git.PathsToSlash()
	}
}

func (p *PackageInfo) PathsFromSlash() {
	if p == nil {
		return
	}
	p.pathsComposite.PathsFromSlash()
	if p.Git != nil {
		p.Git.PathsFromSlash()
	}
	for _, dep := range p.DepsGit {
		dep.Git.PathsFromSlash()
	}
}
