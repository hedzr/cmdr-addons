//go:build linux
// +build linux

package service

import (
	"context"
	"os"
	"testing"

	"github.com/hedzr/is/dir"
)

func TestServiceFile(t *testing.T) {
	ctx := context.Background()
	file := "/tmp/1.service"
	config := &Config{
		Name:           "123",
		DisplayName:    "123 Tool",
		Description:    "123 desc",
		WorkDir:        ".",
		Executable:     "/bin/sh",
		ArgsForInstall: nil,
		Env:            nil,
		RunAs:          "",
		Dependencies:   nil,
		Entity:         nil,

		TempDir: os.TempDir(),
	}
	err := createServiceFile(ctx, config, file)
	if err != nil {
		t.Fatal(err)
	}

	tstfile := "./testdata/123.service"
	if dir.FileExists(tstfile) {
		gen, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		tst, err := os.ReadFile(tstfile)
		if err != nil {
			t.Fatal(err)
		}

		if string(tst) != string(gen) {
			t.Fatalf("generated service file content is not ok.")
		}
	}
}
