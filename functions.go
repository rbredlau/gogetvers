package gogetvers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

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
		tempIterator{"git", []string{"status", "--porcelain"}, &tmp.Status},
		tempIterator{"git", []string{"describe", "--tags", "--abbrev=8", "--always", "--long"}, &tmp.Describe}}
	//
	for _, cmd := range commands {
		fmt.Println("git", cmd.command, cmd.args) //TODO RM
		code, output, err := ExecProgram(path, cmd.command, cmd.args)
		if err == nil && code == 0 {
			output = strings.Trim(output, "\r\n")
			*cmd.target = output
		}
		fmt.Println(*cmd.target) //TODO RM
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
	if tmp.Status != "" {
		rv.Status = tmp.Status
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
