package gogetvers

import (
	"encoding/json"
	"io"
	"os"
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
	ser := &PackageSummary{}
	//
	ser.TargetPackage = pkg.PackageDir
	ser.TargetGit = pkg.GitInfo
	//
	ser.Gits = []*GitInfo{}
	ser.DotDeps = []string{}
	for _, git := range pkg.GitInfos {
		ser.Gits = append(ser.Gits, git)
	}
	//
	for _, name := range pkg.Deps {
		ser.DotDeps = append(ser.DotDeps, name)
	}
	//
	fw, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer fw.Close()
	//
	enc := json.NewEncoder(fw)
	err = enc.Encode(ser)
	if err != nil {
		return err
	}
	//
	return nil
}
