package gogetvers

import (
	"encoding/json"
	"fmt" // TODO RM
	"io"
	"os"
)

type PackageSummary struct {
	TargetPackage string
	TargetGit     *GitInfo
	Gits          []*GitInfo
	DotDeps       []string
}

func Make(sourceDir, outputFile string) error {
	pkg, err := GetPackageInfo(sourceDir)
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
	done := make(chan bool, 2)
	//
	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()
	mw := io.MultiWriter(pw, os.Stdout)
	//
	go func() {
		defer func() { done <- true }()
		enc := json.NewEncoder(mw)
		err = enc.Encode(ser)
		if err != nil {
			fmt.Println("err-> ", err.Error())
		}
	}()
	//
	decoded := &PackageSummary{}
	go func() {
		defer func() { done <- true }()
		dec := json.NewDecoder(pr)
		err = dec.Decode(decoded)
		if err != nil {
			fmt.Println("err-> ", err.Error())
		}
	}()
	//
	select {
	case <-done:
	}
	select {
	case <-done:
	}
	//
	fmt.Printf("decoded-> %#v\n", decoded)
	return nil
}
