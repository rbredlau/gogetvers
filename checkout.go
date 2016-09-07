package gogetvers

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Does a checkout of the manifest at inputFile.
func Checkout(outputDir, inputFile string, statusWriter io.Writer) error {
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
	type gitstat struct {
		git         *GitInfo
		exists      bool
		canCheckout bool
	}
	//
	sw.Write("Checking for dependencies...")
	stats := make(map[*GitInfo]*gitstat)
	for _, v := range ser.Gits {
		stats[v] = &gitstat{git: v, exists: false, canCheckout: false}
	}
	stats[ser.TargetGit] = &gitstat{git: ser.TargetGit, exists: false, canCheckout: false}
	for _, v := range stats {
		chkpath := filepath.Join(outputDir, v.git.HomeDir)
		abs, err := filepath.Abs(chkpath)
		if err != nil {
			return err
		}
		if IsDir(abs) {
			v.exists = true
			currGit, err := GetGitInfo(abs)
			if err != nil {
				return err
			}
			sw.Printf("SWGITSTATS-> %v\n", currGit.Status) //TODO RM
		}
	}
	return errors.New("TODO FINISH ME") //TODO RM
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
