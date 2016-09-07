package gogetvers

import (
	_ "errors"
	"io"
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
	// A local type to record each git, the current status, etc.
	type gitstat struct {
		wanted       *GitInfo // What we want, i.e. what's in the manifest.
		current      *GitInfo // What it currently is.
		dirExists    bool     // If wanted directory exists.
		dir          string
		parentDir    string
		switchBranch bool
		switchHash   bool
		gitClone     bool
	}
	//
	// Gather a summary of operations to perform by checking each wanted git.
	sw.Write("Gathering summary...")
	stats := make(map[string]*gitstat) // Gits by home directory.
	for _, v := range ser.Gits {
		stats[v.HomeDir] = &gitstat{wanted: v, dirExists: false, switchBranch: false, switchHash: false, gitClone: false}
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
		if IsDir(abs) {
			v.dirExists = true
			v.current, err = GetGitInfo(abs)
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
		//		return errors.New("Existing errors prevent further execution.") // TODO PUT ME BACK IN
	}
	sw.Writeln("done")
	//
	// Print the summary
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
					stat.switchBranch = true
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
		}
	}
	sw.Outdent()
	//
	sw.Writeln("Performing checkout...")
	sw.Indent()
	for _, stat := range stats {
		sw.WriteGitInfo(stat.wanted)
		/*
			parentDir := filepath.Join(outputDir, git.ParentDir)
			if !IsDir(parentDir) {
				err = os.MkdirAll(parentDir, 0770)
				if err != nil {
					sw.Error(err)
					return err
				}
			}
			sw.Printf("git clone -b %v %v %v\n", git.Branch, git.OriginUrl, filepath.Base(git.HomeDir))
		*/
		sw.Indent()
		/*
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
		*/
		sw.Writeln("done")
		sw.Outdent()
	}
	sw.Writeln("done")
	sw.Outdent()
	return nil
}
