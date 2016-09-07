package gogetvers

import (
	"encoding/json"
	"io"
	"os"
	"strings"
)

type PackageSummary struct {
	TargetPackage string
	TargetGit     *GitInfo
	Gits          []*GitInfo
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
	pkg, err := GetPackageInfo(sourceDir, sw)
	if err != nil {
		return err
	}
	pkg.ToSlash()
	pkg.StripGoSrcDir()
	sw.Writeln("")
	sw.Writeln("Manifest summary:")
	sw.Indent()
	ser := &PackageSummary{}
	//
	ser.TargetPackage = pkg.PackageDir
	ser.TargetGit = pkg.GitInfo
	sw.Printf("Target package: %v\n", ser.TargetPackage)
	sw.Write("Target ")
	sw.WriteGitInfo(ser.TargetGit)
	//
	ser.Gits = []*GitInfo{}
	ser.DotDeps = []string{}
	sw.Writeln("Dependency gits:")
	sw.Indent()
	for _, git := range pkg.GitInfos {
		ser.Gits = append(ser.Gits, git)
		sw.WriteGitInfo(git)
	}
	sw.Outdent()
	sw.Outdent()
	sw.Writeln("")
	// Warn for untracked stuff.
	if len(pkg.Untrackable) > 0 {
		sw.Warning("The following dependencies are NOT tracked:")
		sw.Indent()
		for _, v := range pkg.Untrackable {
			sw.Writeln(v.Name)
		}
		sw.Outdent()
		sw.Writeln("")
	}
	//
	for _, name := range pkg.Deps {
		ser.DotDeps = append(ser.DotDeps, name)
	}
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
	err = enc.Encode(ser)
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
	sw.WriteGitInfo(ser.TargetGit)
	//
	sw.Writeln("Dependency gits:")
	sw.Indent()
	for _, git := range ser.Gits {
		sw.WriteGitInfo(git)
	}
	sw.Outdent()
	sw.Outdent()
	sw.Writeln("")
	sw.Printf("Dependencies-> %v\n", strings.Join(ser.DotDeps, " "))
	//
	return nil
}
