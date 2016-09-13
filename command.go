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

type FuncCommandOutputProcessor func(output string) string

type Command struct {
	Bin             string
	Args            []string
	Output          string
	ExitCode        int
	OutputProcessor FuncCommandOutputProcessor
}

func NewCommandGitBranch() *Command {
	return NewCommand("git", "branch")
}

func NewCommandGitCheckout(hash string) *Command {
	return NewCommand("git", "checkout", hash)
}

func NewCommandGitClone(branch, origin, outputDir string) *Command {
	return NewCommand("git", "clone", "-b", branch, origin, outputDir)
}

func NewCommandGitDescribe() *Command {
	return NewCommand("git", "describe", "--tags", "--abbrev=8", "--always", "--long")
}

func NewCommandGitHash() *Command {
	return NewCommand("git", "rev-parse", "HEAD")
}

func NewCommandGitOrigin() *Command {
	return NewCommand("git", "config", "--get", "remote.origin.url")
}

func NewCommandGitStatus() *Command {
	return NewCommand("git", "status", "--porcelain")
}

func NewCommandGoFmt(file ...string) *Command {
	return NewCommand("go", append([]string{"fmt"}, file...)...)
}

func NewCommandGoList() *Command {
	return NewCommand("go", "list")
}

func NewCommandGoListDeps() *Command {
	rv := NewCommand("go", "list", "-f", "{{.Deps}}")
	rv.OutputProcessor = func(output string) string {
		return strings.Replace(strings.Replace(output, "[", "", 1), "]", "", 1)
	}
	return rv
}

func NewCommand(bin string, args ...string) *Command {
	rv := &Command{Bin: bin, Args: []string{}, ExitCode: -1}
	for _, v := range args {
		rv.Args = append(rv.Args, v)
	}
	return rv
}

func (cmd *Command) String() string {
	if cmd == nil {
		return ""
	}
	return strings.Join(append([]string{cmd.Bin}, cmd.Args...), " ")
}

func (cmd *Command) Exec(chdir string) error {
	if cmd == nil {
		return errors.New("nil receiver")
	}
	// Exit code and standard output.
	cmd.Output = ""
	cmd.ExitCode = -1
	// Catch errors
	var err error
	// Done channel tells us when command is done.
	done := make(chan error, 1)
	// Create command.
	runme := exec.Command(cmd.Bin, cmd.Args...)
	// Standard output handler
	stdoutRet := make(chan bool, 1)
	defer func() {
		// Trim output but not until after the stdout goroutine is done.
		select {
		case <-stdoutRet:
		}
		cmd.Output = strings.TrimSpace(cmd.Output)
		if cmd.OutputProcessor != nil {
			cmd.Output = cmd.OutputProcessor(cmd.Output)
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
					cmd.Output = cmd.Output + string(bytes.TrimRight(dat, "\x00"))
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
		cmd.ExitCode = 0
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					cmd.ExitCode = status.ExitStatus()
					err = nil
				}
			}
		}
	}
	if err != nil {
		return err
	}
	if cmd.ExitCode != 0 {
		return errors.New(fmt.Sprintf("%v returns %v", cmd.String(), cmd.ExitCode))
	}

	return nil
}
