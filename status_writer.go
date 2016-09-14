package gogetvers

import (
	"fmt"
	"io"
	"strings"
)

// A utility type for printing status information to an io.Writer.
type StatusWriter struct {
	Writer      io.Writer
	IndentLevel int
}

// Like fmt.Printf()
func (st *StatusWriter) Printf(fmtstr string, args ...interface{}) {
	st.Write(fmt.Sprintf(fmtstr, args...))
}

// Writes a string to the writer.
func (st *StatusWriter) Write(str string) {
	if st == nil {
		return
	}
	io.WriteString(st.Writer, strings.Repeat(" ", st.IndentLevel)+str)
}

// Writes a summary for a Git type to the writer.
func (st *StatusWriter) WriteGit(gi *Git) {
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

// Writes the string and appends a newline.
func (st *StatusWriter) Writeln(str string) {
	st.Write(str + "\n")
}

// Writes an error with ERROR prefix to the writer.
func (st *StatusWriter) Error(err error) {
	st.Printf("ERROR: %v\n", err.Error())
}

// Writes string with a WARNING prefix.
func (st *StatusWriter) Warning(str string) {
	st.Writeln("WARNING: " + str)
}

// After calling Indent() new writes will be indented by 4 spaces; keep
// calling Indent() to perform nested indenting.
func (st *StatusWriter) Indent() {
	if st == nil {
		return
	}
	st.IndentLevel = st.IndentLevel + 4
}

// Undoes the most recent level of Indent().
func (st *StatusWriter) Outdent() {
	if st == nil {
		return
	}
	st.IndentLevel = st.IndentLevel - 4
	if st.IndentLevel < 0 {
		st.IndentLevel = 0
	}
}
