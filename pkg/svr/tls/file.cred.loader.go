// Copyright © 2020 Hedzr Yeh.

package tls

import (
	"bufio"
	"io"
	"os"
	"regexp"
)

type fileCredLoader struct {
	file    string
	inFile  *os.File
	scanner *bufio.Scanner
	re      *regexp.Regexp
}

func newFileCredLoader() *fileCredLoader {
	return &fileCredLoader{
		file: "ci/credentials.txt",
	}
}

func (f *fileCredLoader) Close() (err error) {
	if f.inFile != nil {
		if f.scanner != nil {
			f.scanner = nil
		}
		if f.re != nil {
			f.re = nil
		}
		err = f.inFile.Close()
		f.inFile = nil
	}
	return
}

func (f *fileCredLoader) Next() (user, hash string, err error) {
	if f.inFile == nil {
		f.inFile, err = os.Open(f.file)
		if err != nil {
			return
		}
		f.scanner = bufio.NewScanner(f.inFile)
		f.re, err = regexp.Compile(`(.+)[ \t]+(.+)`)
		if err != nil {
			return
		}
	}
	// defer f.inFile.Close()

	if f.scanner.Scan() {
		parts := f.re.Split(f.scanner.Text(), -1)
		user, hash = parts[0], parts[1]
	} else {
		err = io.EOF
		f.Close()
	}
	return
}
