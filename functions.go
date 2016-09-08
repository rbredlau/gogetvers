package gogetvers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Starts at path and works upwards looking for .git directory.
// Stops when it reaches stopDir and returns an error.
func GetGitPath(path, stopDir string) (string, error) {
	if path == "" || !IsDir(path) {
		return "", errors.New(fmt.Sprintf("Not a path @ %v", path))
	}
	if stopDir == "" || !IsDir(stopDir) {
		return "", errors.New(fmt.Sprintf("Not a path @ %v", stopDir))
	}
	if stopDir == path {
		return "", errors.New(fmt.Sprintf("Search for git reached stopDir"))
	}
	try := filepath.Join(path, ".git")
	if IsDir(try) {
		abs, err := filepath.Abs(try)
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	return GetGitPath(filepath.Dir(path), stopDir)
}

// Get dependency info for a depency.
func GetDependencyInfo(pkg *PackageInfo, status *StatusWriter) error {
	if pkg == nil {
		return errors.New("pkg is nil")
	}
	for _, v := range pkg.Deps {
		status.Printf("Dependency -> %v\n", v)
		status.Indent()
		info := newDependencyInfo(true, false, v, filepath.FromSlash(filepath.Join(pkg.GoSrcDir, v)))
		if IsDir(info.DepDir) {
			info.IsGo = false
			gitDir, err := GetGitPath(info.DepDir, pkg.GoSrcDir)
			if err == nil && gitDir != "" {
				status.Writeln("git repository")
				info.IsGit = true
				info.GitDir = gitDir
				status.Indent()
				if _, ok := pkg.GitDirs[gitDir]; ok {
					pkg.GitDirs[gitDir] = append(pkg.GitDirs[gitDir], info)
					status.Writeln("previously discovered")
				} else {
					pkg.GitDirs[gitDir] = []*DependencyInfo{info}
					pkg.Gits[gitDir], _ = NewGit(filepath.Dir(gitDir))
					status.WriteGit(pkg.Gits[gitDir])
				}
				status.Outdent()
			}
		} else {
			info.DepDir = ""
			status.Writeln("golang standard package")

		}
		pkg.DepInfo[v] = info
		if !info.IsGo && !info.IsGit {
			status.Warning("Not a standard package and not a git repository; untrackable.")
			pkg.Untrackable[v] = info
		}
		status.Outdent()
	}
	return nil
}

// For a given packagePath returns a PackageInfo type.
func GetPackageInfo(packagePath string, status *StatusWriter) (*PackageInfo, error) {
	var rv *PackageInfo
	// Get absolute path of sourceDir
	status.Printf("Get package info for %v\n", packagePath)
	packageDir, err := filepath.Abs(packagePath)
	if err != nil {
		status.Error(err)
		return nil, err
	}

	// Get info for package at sourceDir
	golist := newCommandGoList()
	err = golist.exec(packageDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	status.Printf("%v -> %v\n", golist.String(), golist.output)

	// We can now deduce go-src path.
	goSrcDir := strings.Replace(filepath.ToSlash(packageDir), golist.output, "", -1)
	goSrcDir, err = filepath.Abs(goSrcDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	goSrcDir = strings.TrimRight(goSrcDir, "\\/")
	status.Printf("go-src path -> %v\n", goSrcDir)

	// Get dependency information for package at sourceDir
	golistdeps := newCommandGoListDeps()
	err = golistdeps.exec(packageDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	rv = &PackageInfo{PackageDir: packageDir, GoSrcDir: goSrcDir,
		Deps: []string{}, DepInfo: make(map[string]*DependencyInfo),
		GitDirs:     make(map[string][]*DependencyInfo),
		Gits:        make(map[string]*Git),
		Untrackable: make(map[string]*DependencyInfo)}
	lines := strings.Split(golistdeps.output, " ")
	status.Indent()
	for _, v := range lines {
		status.Printf("dependency -> %v\n", v)
		rv.Deps = append(rv.Deps, v)
	}
	status.Outdent()
	//
	status.Writeln("")
	status.Write("Getting git-info for target package...")
	rv.GitDir, err = GetGitPath(packageDir, goSrcDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	rv.Git, err = NewGit(filepath.Dir(rv.GitDir))
	if err != nil {
		status.Error(err)
		return nil, err
	}
	status.Writeln("done")
	status.WriteGit(rv.Git)
	//
	status.Writeln("")
	status.Indent()
	status.Writeln("Getting dependency information...")
	err = GetDependencyInfo(rv, status)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	status.Outdent()
	return rv, nil
}

// Opens the input file and decodes the manifest.
func LoadManifest(inputFile string) (*PackageSummary, error) {
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
	summary := &PackageSummary{}
	err = dec.Decode(summary)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

// Determines if path is a file; returns true if it is.
func IsFile(path string) bool {
	if len(path) == 0 {
		return false
	}
	finfo, err := os.Stat(path)
	return err == nil && !finfo.IsDir()

}

// Determines if path is a directory; returns true if it is.
func IsDir(path string) bool {
	if len(path) == 0 {
		return false
	}
	finfo, err := os.Stat(path)
	return err == nil && finfo.IsDir()
}

// Attempt to read file and return the data.
func ReadFile(path string) (rvdata []byte, rverr error) {
	if !IsFile(path) {
		rverr = errors.New(fmt.Sprintf("Not a file @ %v", path))
		return
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		rverr = err
		return
	}
	rvdata = data
	return
}

// Attempt to write date to file.
func WriteFile(path string, data []byte, overwrite bool) (rverr error) {
	var err error
	if IsDir(path) {
		rverr = errors.New(fmt.Sprintf("path is a directory @ %v", path))
		return
	}
	if IsFile(path) && !overwrite {
		fil, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0660)
		if err != nil {
			rverr = err
			return
		}
		defer fil.Close()
		if _, err = fil.Write(data); err != nil {
			rverr = err
			return
		}
	} else {
		err = ioutil.WriteFile(path, data, 0660)
		if err != nil {
			rverr = err
			return
		}
	}
	return
}
