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
	//
	// Combine all gits by their home directory; this is in case
	// TargetGit is the same git as a dependency.
	gits := make(map[string]*Git)
	for _, v := range ser.Gits {
		gits[v.HomeDir] = v
	}
	// Check each directory and see if already exists; return error if it does.
	gits[ser.TargetGit.HomeDir] = ser.TargetGit
	for v, _ := range gits {
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
	// Now clone each git into the proper directory.
	sw.Writeln("Cloning gits...")
	sw.Indent()
	for _, git := range gits {
		sw.WriteGit(git)
		parentDir := filepath.Join(outputDir, git.ParentDir)
		if !IsDir(parentDir) {
			err = os.MkdirAll(parentDir, 0770)
			if err != nil {
				sw.Error(err)
				return err
			}
		}
		cmd := newCommandGitClone("master", git.OriginUrl, filepath.Base(git.HomeDir))
		sw.Writeln(cmd.String())
		sw.Indent()
		err = cmd.exec(parentDir)
		if err != nil {
			sw.Error(err)
			return err
		}
		// TODO Checkout appropriate hash; simply cloning is not enough.
		sw.Writeln("done")
		sw.Outdent()
	}
	sw.Writeln("done")
	sw.Outdent()
	return nil
}
