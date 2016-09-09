package gogetvers

import (
	"encoding/json"
	"io"
	"os"
	"strings"
)

type PackageSummary struct {
	TargetPackage string
	TargetGit     *Git
	Gits          []*Git
	DotDeps       []string
}

// Inspects the go package located at sourceDir and creates
// a manifest file at outputFile location.  If statusWriter
// is non-nil then output will be written there.
func Make(sourceDir, outputFile string, statusWriter io.Writer) error {
	var sw *StatusWriter
	if statusWriter != nil {
		sw = &StatusWriter{Writer: statusWriter}
	}
	pi, err := getPackageInfo(sourceDir, sw)
	if err != nil {
		sw.Error(err)
		return err
	}
	//
	pi.StripPathPrefix(pi.RootDir)
	sw.Writeln(pi.getSummary())
	//
	sw.Printf("Writing output to %v\n", outputFile)
	sw.Indent()
	fw, err := os.Create(outputFile)
	if err != nil {
		sw.Error(err)
		return err
	}
	defer fw.Close()
	//
	enc := json.NewEncoder(fw)
	err = enc.Encode(pi)
	if err != nil {
		sw.Error(err)
		return err
	}
	sw.Writeln("done")
	sw.Outdent()
	sw.Writeln("")
	//
	return nil
}

// Reads a manifest file and prints summary information.
func Print(sourceDir, inputFile string, statusWriter io.Writer) error {
	var sw *StatusWriter
	if statusWriter != nil {
		sw = &StatusWriter{Writer: statusWriter}
	}
	ser, err := LoadManifest(inputFile)
	if err != nil {
		return err
	}
	//
	sw.Writeln(inputFile)
	sw.Writeln("Manifest summary:")
	sw.Indent()
	//
	sw.Printf("Target package: %v\n", ser.TargetPackage)
	sw.Write("Target ")
	sw.WriteGit(ser.TargetGit)
	//
	sw.Writeln("Dependency gits:")
	sw.Indent()
	for _, git := range ser.Gits {
		sw.WriteGit(git)
	}
	sw.Outdent()
	sw.Outdent()
	sw.Writeln("")
	sw.Printf("Dependencies-> %v\n", strings.Join(ser.DotDeps, " "))
	//
	return nil
}
