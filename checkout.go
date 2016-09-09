package gogetvers

import (
	"io"
	/* TODO
	fs "broadlux/fileSystem"
	"errors"
	"os"
	"path/filepath"
	*/)

// Does a checkout of the manifest at inputFile.
func Checkout(outputDir, inputFile string, statusWriter io.Writer) error {
	return nil // TODO
	/*
		var sw *StatusWriter
		if statusWriter != nil {
			sw = &StatusWriter{Writer: statusWriter}
		}
		//
		ser, err := loadPackageInfoFile(inputFile)
		if err != nil {
			return err
		}
		//
		// A local type to record each git, the current status, etc.
		type gitstat struct {
			wanted     *Git // What we want, i.e. what's in the manifest.
			current    *Git // What it currently is.
			dirExists  bool // If wanted directory exists.
			dir        string
			parentDir  string
			switchHash bool
			gitClone   bool
		}
		//
		// Gather a summary of operations to perform by checking each wanted git.
		sw.Write("Gathering summary...")
		stats := make(map[string]*gitstat) // Gits by home directory.
		for _, v := range ser.Gits {
			stats[v.HomeDir] = &gitstat{wanted: v, dirExists: false, switchHash: false, gitClone: false}
		}
		stats[ser.TargetGit.HomeDir] = &gitstat{wanted: ser.TargetGit}
		gitsWithMods := []string{} // gits with local modifications
		dirsNotGits := []string{}  // directories that exist but are not a git clone
		for _, v := range stats {
			chkpath := filepath.Join(outputDir, v.wanted.HomeDir)
			abs, err := filepath.Abs(chkpath)
			if err != nil {
				return err
			}
			if fs.IsDir(abs) {
				v.dirExists = true
				v.current, err = NewGit(abs)
				if err == nil && v.current != nil {
					if v.current.Status != "" {
						gitsWithMods = append(gitsWithMods, abs)
					}
				} else {
					v.current = nil
					dirsNotGits = append(dirsNotGits, abs)
				}
			} else {
				v.dirExists = false
			}
			v.dir = abs
			v.parentDir = filepath.Dir(abs)
		}
		//
		// If either gitsWithMods or dirsNotGits is non-empty then abort.
		if len(gitsWithMods) > 0 || len(dirsNotGits) > 0 {
			if len(gitsWithMods) > 0 {
				sw.Writeln("The following gits have local modifications:")
				sw.Indent()
				for _, v := range gitsWithMods {
					sw.Writeln(v)
				}
				sw.Outdent()
			}
			if len(dirsNotGits) > 0 {
				sw.Writeln("The following directories already exist but are not a git clone:")
				sw.Indent()
				for _, v := range dirsNotGits {
					sw.Writeln(v)
				}
				sw.Outdent()
			}
			return errors.New("Existing errors prevent further execution.")
		}
		sw.Writeln("done")
		//
		// Print the summary
		sw.Writeln("")
		sw.Writeln("What will be done:")
		sw.Indent()
		for _, stat := range stats {
			if stat.dirExists {
				sw.Writeln(stat.dir)
				sw.Indent()
				if stat.wanted.Branch == stat.current.Branch && stat.wanted.Hash == stat.current.Hash {
					sw.Writeln("branch and hash are current; nothing to do")
				} else {
					if stat.wanted.Branch != stat.current.Branch {
						sw.Printf("switch branch from %v to %v\n", stat.current.Branch, stat.wanted.Branch)
					}
					if stat.wanted.Hash != stat.current.Hash {
						sw.Printf("switch hash from %v to %v\n", stat.current.Hash, stat.wanted.Hash)
						stat.switchHash = true
					}
				}
				sw.Outdent()
			} else {
				sw.Printf("%v will be cloned\n", stat.dir)
				stat.gitClone = true
				stat.switchHash = true
			}
		}
		sw.Outdent()
		//
		sw.Writeln("")
		sw.Writeln("Performing work...")
		sw.Indent()
		for _, stat := range stats {
			sw.Writeln(stat.dir)
			if !stat.gitClone && !stat.switchHash {
				sw.Indent()
				sw.Writeln("skipping")
				sw.Outdent()
				continue
			}
			if !fs.IsDir(stat.parentDir) {
				err = os.MkdirAll(stat.parentDir, 0770)
				if err != nil {
					sw.Error(err)
					return err
				}
			}
			if stat.gitClone {
				cmd := newCommandGitClone("master", stat.wanted.OriginUrl, filepath.Base(stat.wanted.HomeDir))
				sw.Writeln(cmd.String())
				err = cmd.exec(stat.parentDir)
				if err != nil {
					sw.Error(err)
					return err
				}
			}
			if stat.switchHash {
				cmd := newCommandGitCheckout(stat.wanted.Hash)
				sw.Writeln(cmd.String())
				err = cmd.exec(stat.dir)
				if err != nil {
					sw.Error(err)
					return err
				}
			}
			sw.Indent()
			sw.Writeln("done")
			sw.Outdent()
		}
		sw.Outdent()
		sw.Writeln("done")
		return nil
	*/
}
