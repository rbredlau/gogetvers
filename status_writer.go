package gogetvers

import (
	"fmt"
	"io"
	"strings"
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
