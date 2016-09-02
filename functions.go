package gogetvers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type StatusWriter struct {
	Writer      io.Writer
	IndentLevel int
}

func (st *StatusWriter) Printf(fmtstr string, args ...interface{}) {
	st.Write(fmt.Sprintf(fmtstr, args...))
}

func (st *StatusWriter) Write(str string) {
	if st == nil {
		return
	}
	io.WriteString(st.Writer, strings.Repeat(" ", st.IndentLevel)+str)
}

func (st *StatusWriter) WriteGitInfo(gi *GitInfo) {
	if st == nil {
		return
	}
	st.Write("git-info -> ")
	if gi == nil {
		st.Writeln("nil")
	} else {
		st.Writeln("")
		st.Printf("Home -> %v\n", gi.HomeDir)
		st.Indent()
		st.Printf("Hash -> %v\n", gi.Hash)
		st.Printf("Branch -> %v\n", gi.Branch)
		st.Printf("Origin -> %v\n", gi.OriginUrl)
		st.Printf("Describe -> %v\n", gi.Describe)
		st.Outdent()
	}
}

func (st *StatusWriter) Writeln(str string) {
	st.Write(str + "\n")
}

func (st *StatusWriter) Error(err error) {
	st.Printf("ERROR: %v\n", err.Error())
}

func (st *StatusWriter) Warning(str string) {
	st.Writeln("WARNING: " + str)
}

func (st *StatusWriter) Indent() {
	if st == nil {
		return
	}
	st.IndentLevel = st.IndentLevel + 4
}

func (st *StatusWriter) Outdent() {
	if st == nil {
		return
	}
	st.IndentLevel = st.IndentLevel - 4
	if st.IndentLevel < 0 {
		st.IndentLevel = 0
	}
}

type PackageInfo struct {
	PackageDir  string                     // Package source directory; absolute path.
	GitDir      string                     // Path to .git for package.
	GitInfo     *GitInfo                   // Git info for package.
	Deps        []string                   // List of package dependencies.
	GoSrcDir    string                     // Absolute path of Go src that contains SourceDir.
	DepInfo     map[string]*DependencyInfo // Map of dependency info.
	GitDirs     map[string][]*DependencyInfo
	GitInfos    map[string]*GitInfo
	Untrackable map[string]*DependencyInfo
}

func (p *PackageInfo) StripGoSrcDir() {
	if p == nil {
		return
	}
	p.PackageDir = strings.TrimLeft(strings.Replace(p.PackageDir, p.GoSrcDir, "", -1), "\\/")
	p.GitDir = strings.TrimLeft(strings.Replace(p.GitDir, p.GoSrcDir, "", -1), "\\/")
	if p.GitInfo != nil {
		p.GitInfo.StripGoSrcDir(p.GoSrcDir)
	}
	for _, v := range p.DepInfo {
		v.StripGoSrcDir(p.GoSrcDir)
	}
	for _, v := range p.GitInfos {
		v.StripGoSrcDir(p.GoSrcDir)
	}
}

func (p *PackageInfo) ToSlash() {
	if p == nil {
		return
	}
	p.PackageDir = filepath.ToSlash(p.PackageDir)
	p.GitDir = filepath.ToSlash(p.GitDir)
	p.GoSrcDir = filepath.ToSlash(p.GoSrcDir)
	if p.GitInfo != nil {
		p.GitInfo.ToSlash()
	}
	for _, v := range p.DepInfo {
		v.ToSlash()
	}
	for _, v := range p.GitInfos {
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
	if p.GitInfo != nil {
		p.GitInfo.FromSlash()
	}
	for _, v := range p.DepInfo {
		v.FromSlash()
	}
	for _, v := range p.GitInfos {
		v.FromSlash()
	}
}

type DependencyInfo struct {
	IsGo   bool   // True if not in GoSrcDir as a path.
	IsGit  bool   // True if .git info was found.
	Name   string // Name according to: go list -f {{.Deps}} from the parent package.
	DepDir string // Path to dependency.
	GitDir string // The .git directory.
}

func (d *DependencyInfo) StripGoSrcDir(path string) {
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

type GitInfo struct {
	HomeDir   string
	ParentDir string
	Branch    string
	Hash      string
	OriginUrl string
	Describe  string
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

// Returns git info for path.
func GetGitInfo(path string) (*GitInfo, error) {
	if !IsDir(path) {
		return nil, errors.New(fmt.Sprintf("not a path @ %v", path))
	}
	rv := &GitInfo{HomeDir: path, ParentDir: filepath.Dir(path)}
	tmp := &GitInfo{}
	type tempIterator struct {
		command string
		args    []string
		target  *string
	}
	commands := []tempIterator{
		tempIterator{"git", []string{"branch"}, &tmp.Branch},
		tempIterator{"git", []string{"config", "--get", "remote.origin.url"}, &tmp.OriginUrl},
		tempIterator{"git", []string{"rev-parse", "HEAD"}, &tmp.Hash},
		tempIterator{"git", []string{"describe", "--tags", "--abbrev=8", "--always", "--long"}, &tmp.Describe}}
	//
	for _, cmd := range commands {
		code, output, err := ExecProgram(path, cmd.command, cmd.args)
		if err == nil && code == 0 {
			output = strings.Trim(output, "\r\n")
			*cmd.target = output
		}
	}
	//
	if tmp.Branch != "" {
		pieces := strings.Split(tmp.Branch, "\n")
		for _, v := range pieces {
			v = strings.Trim(v, "\r\n ")
			if v[0] == '*' {
				pieces := strings.Split(v, " ")
				rv.Branch = strings.Trim(pieces[1], " ")
			}
			break
		}
	}
	if tmp.Hash != "" {
		rv.Hash = tmp.Hash
	}
	if tmp.OriginUrl != "" {
		rv.OriginUrl = tmp.OriginUrl
	}
	if tmp.Describe != "" {
		rv.Describe = tmp.Describe
	}
	return rv, nil
}

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
	//
	done := make(chan bool, 1)
	// Do scan.
	go func() {
		defer func() { done <- true }()
		for _, v := range pkg.Deps {
			status.Printf("Dependency -> %v\n", v)
			status.Indent()
			info := &DependencyInfo{IsGo: true, IsGit: false, Name: v, DepDir: filepath.FromSlash(filepath.Join(pkg.GoSrcDir, v))}
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
						pkg.GitInfos[gitDir], _ = GetGitInfo(filepath.Dir(gitDir))
						status.WriteGitInfo(pkg.GitInfos[gitDir])
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
	}()
	// Wait for scan to complete
	select {
	case <-done:
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
	code, output, err := ExecProgram(packageDir, "go", []string{"list"})
	if err != nil {
		status.Error(err)
		return nil, err
	}
	if code != 0 {
		err = errors.New(fmt.Sprintf("go list returns-> %v", code))
		status.Error(err)
		return nil, err
	}
	output = strings.Trim(output, "\r\n")
	status.Printf("go list yields -> %v\n", output)

	// We can now deduce go-src path.
	goSrcDir := strings.Replace(filepath.ToSlash(packageDir), output, "", -1)
	goSrcDir, err = filepath.Abs(goSrcDir)
	if err != nil {
		status.Error(err)
		return nil, err
	}
	goSrcDir = strings.TrimRight(goSrcDir, "\\/")
	status.Printf("go-src path -> %v\n", goSrcDir)

	// Get dependency information for package at sourceDir
	code, output, err = ExecProgram(packageDir, "go", []string{"list", "-f", "{{.Deps}}"})
	if err != nil {
		status.Error(err)
		return nil, err
	}
	if code != 0 {
		err = errors.New(fmt.Sprintf("go list returns-> %v", code))
		status.Error(err)
		return rv, err
	}
	rv = &PackageInfo{PackageDir: packageDir, GoSrcDir: goSrcDir,
		Deps: []string{}, DepInfo: make(map[string]*DependencyInfo),
		GitDirs:     make(map[string][]*DependencyInfo),
		GitInfos:    make(map[string]*GitInfo),
		Untrackable: make(map[string]*DependencyInfo)}
	output = strings.Trim(output, "\r\n[]")
	lines := strings.Split(output, " ")
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
	rv.GitInfo, err = GetGitInfo(filepath.Dir(rv.GitDir))
	if err != nil {
		status.Error(err)
		return nil, err
	}
	status.Writeln("done")
	status.WriteGitInfo(rv.GitInfo)
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

// Executes "binary" with "args" in path and returns standard output, error code
// and any error.
func ExecProgram(path, binary string, args []string) (int, string, error) {
	// Exit code and standard output.
	output := ""
	exitCode := int(-1)
	// Catch errors
	var err error
	// Done channel tells us when command is done.
	done := make(chan error, 1)
	// Create command.
	cmd := exec.Command(binary, args...)
	// Standard output handler
	stdoutDone := make(chan bool, 1)
	defer func() { stdoutDone <- true }()
	stdoutRdr, err := cmd.StdoutPipe()
	if err != nil {
		return exitCode, output, err
	}
	go func() {
		for {
			select {
			case <-stdoutDone:
				return
			default:
				dat := make([]byte, 256)
				nn, _ := stdoutRdr.Read(dat)
				if nn > 0 {
					output = output + string(bytes.TrimRight(dat, "\x00"))
				} else {
					time.Sleep(300 * time.Millisecond)
				}
			}
		}
	}()
	// Start command
	started := make(chan bool, 1)
	go func() {
		cw, err := os.Getwd()
		if err != nil {
			return
		}
		defer os.Chdir(cw)
		os.Chdir(path)
		err = cmd.Start()
		started <- true
		if err != nil {
			return
		}
	}()

	go func() {
		select {
		case <-started:
		}
		done <- cmd.Wait()
	}()
	select {
	case err = <-done:
		exitCode = 0
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					exitCode = status.ExitStatus()
					err = nil
				}
			}
		}
	}
	if err != nil {
		return exitCode, output, err
	}

	return exitCode, output, nil
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
