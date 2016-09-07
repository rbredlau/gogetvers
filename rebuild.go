package gogetvers

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Rebuilds the information contained within input file in outputDir.
func Rebuild(outputDir, inputFile string, statusWriter io.Writer) error {
	var sw *StatusWriter
	if statusWriter != nil {
		sw = &StatusWriter{Writer: statusWriter}
	}
	//
	ser, err := LoadManifest(inputFile)
	if err != nil {
		return err
	}
	//
	sw.Write("Checking for directory collisions...")
	gitDirs := make(map[string]bool)
	for _, v := range ser.Gits {
		gitDirs[v.HomeDir] = true
	}
	gitDirs[ser.TargetGit.HomeDir] = true
	for v, _ := range gitDirs {
		chkpath := filepath.Join(outputDir, v)
		abs, err := filepath.Abs(chkpath)
		if err != nil {
			return err
		}
		if IsDir(abs) {
			err := errors.New(fmt.Sprintf("directory already exists @ %v", abs))
			sw.Error(err)
			return err
		}
	}
	sw.Writeln("done")
	//
	sw.Writeln("Cloning gits...")
	sw.Indent()
	for _, git := range ser.Gits {
		sw.WriteGitInfo(git)
		parentDir := filepath.Join(outputDir, git.ParentDir)
		if !IsDir(parentDir) {
			err = os.MkdirAll(parentDir, 0770)
			if err != nil {
				sw.Error(err)
				return err
			}
		}
		sw.Printf("git clone -b %v %v %v\n", git.Branch, git.OriginUrl, filepath.Base(git.HomeDir))
		sw.Indent()
		code, _, err := ExecProgram(parentDir, "git", []string{"clone", "-b", git.Branch, git.OriginUrl, filepath.Base(git.HomeDir)})
		if err != nil {
			sw.Error(err)
			return err
		}
		if code != 0 {
			err := errors.New(fmt.Sprintf("git clone returns -> %v", code))
			sw.Error(err)
			return err
		}
		sw.Writeln("done")
		sw.Outdent()
	}
	sw.Writeln("done")
	sw.Outdent()
	return nil
}
