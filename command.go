package gogetvers

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type command struct {
	bin      string
	args     []string
	output   string
	exitCode int
}

/* TODO RM
tempIterator{"git", []string{"branch"}, &rv.Branch},
tempIterator{"git", []string{"config", "--get", "remote.origin.url"}, &rv.OriginUrl},
tempIterator{"git", []string{"rev-parse", "HEAD"}, &rv.Hash},
tempIterator{"git", []string{"status", "--porcelain"}, &rv.Status},
tempIterator{"git", []string{"describe", "--tags", "--abbrev=8", "--always", "--long"}, &rv.Describe}}
*/

func newCommandGitBranch() *command {
	return newCommand("git", "branch")
}

func newCommandGitOrigin() *command {
	return newCommand("git", "config", "--get", "remote.origin.url")
}

func newCommandGitHash() *command {
	return newCommand("git", "rev-parse", "HEAD")
}

func newCommandGitStatus() *command {
	return newCommand("git", "status", "--porcelain")
}

func newCommandGitDescribe() *command {
	return newCommand("git", "describe", "--tags", "--abbrev=8", "--always", "--long")
}

func newCommand(bin string, args ...string) *command {
	rv := &command{bin: bin, args: []string{}, exitCode: -1}
	for _, v := range args {
		rv.args = append(rv.args, v)
	}
	return rv
}

func (cmd *command) exec(chdir string) error {
	if cmd == nil {
		return errors.New("nil receiver")
	}
	// Exit code and standard output.
	cmd.output = ""
	cmd.exitCode = -1
	// Catch errors
	var err error
	// Done channel tells us when command is done.
	done := make(chan error, 1)
	// Create command.
	runme := exec.Command(cmd.bin, cmd.args...)
	// Standard output handler
	stdoutDone := make(chan bool, 1)
	defer func() { stdoutDone <- true }()
	stdoutRdr, err := runme.StdoutPipe()
	if err != nil {
		return err
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
					cmd.output = cmd.output + string(bytes.TrimRight(dat, "\x00"))
				} else {
					time.Sleep(300 * time.Millisecond)
				}
			}
		}
	}()
	// Start command
	started := make(chan bool, 1)
	go func() {
		if chdir != "" {
			cw, err := os.Getwd()
			if err != nil {
				return
			}
			defer os.Chdir(cw)
			os.Chdir(chdir)
		}
		err = runme.Start()
		started <- true
		if err != nil {
			return
		}
	}()

	go func() {
		select {
		case <-started:
		}
		done <- runme.Wait()
	}()
	select {
	case err = <-done:
		cmd.exitCode = 0
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					cmd.exitCode = status.ExitStatus()
					err = nil
				}
			}
		}
	}
	if err != nil {
		return err
	}
	if cmd.exitCode != 0 {
		return errors.New(fmt.Sprintf("%v %v returns %v", cmd.bin, strings.Join(cmd.args, " "), cmd.exitCode))
	}

	return nil
}
