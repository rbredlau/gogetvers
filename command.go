package gogetvers

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type funcCommandOutputProcessor func(output string) string

type command struct {
	bin             string
	args            []string
	output          string
	exitCode        int
	outputProcessor funcCommandOutputProcessor
}

func newCommandGitBranch() *command {
	return newCommand("git", "branch")
}

func newCommandGitCheckout(hash string) *command {
	return newCommand("git", "checkout", hash)
}

func newCommandGitClone(branch, origin, outputDir string) *command {
	return newCommand("git", "clone", "-b", branch, origin, outputDir)
}

func newCommandGitDescribe() *command {
	return newCommand("git", "describe", "--tags", "--abbrev=8", "--always", "--long")
}

func newCommandGitHash() *command {
	return newCommand("git", "rev-parse", "HEAD")
}

func newCommandGitOrigin() *command {
	return newCommand("git", "config", "--get", "remote.origin.url")
}

func newCommandGitStatus() *command {
	return newCommand("git", "status", "--porcelain")
}

func newCommandGoFmt(file ...string) *command {
	args := append([]string{"fmt"}, file...)
	return newCommand("go", args...)
}

func newCommandGoList() *command {
	return newCommand("go", "list")
}

func newCommandGoListDeps() *command {
	rv := newCommand("go", "list", "-f", "{{.Deps}}")
	rv.outputProcessor = func(output string) string {
		return strings.Replace(strings.Replace(output, "[", "", 1), "]", "", 1)
	}
	return rv
}

func newCommand(bin string, args ...string) *command {
	rv := &command{bin: bin, args: []string{}, exitCode: -1}
	for _, v := range args {
		rv.args = append(rv.args, v)
	}
	return rv
}

func (cmd *command) String() string {
	if cmd == nil {
		return ""
	}
	return strings.Join(append([]string{cmd.bin}, cmd.args...), " ")
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
	stdoutRet := make(chan bool, 1)
	defer func() {
		// Trim output but not until after the stdout goroutine is done.
		select {
		case <-stdoutRet:
		}
		cmd.output = strings.TrimSpace(cmd.output)
		if cmd.outputProcessor != nil {
			cmd.output = cmd.outputProcessor(cmd.output)
		}
	}()
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
				stdoutRet <- true
				return
			default:
				dat := make([]byte, 256)
				nn, _ := stdoutRdr.Read(dat)
				if nn > 0 {
					cmd.output = cmd.output + string(bytes.TrimRight(dat, "\x00"))
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
		return errors.New(fmt.Sprintf("%v returns %v", cmd.String(), cmd.exitCode))
	}

	return nil
}
