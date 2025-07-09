//go:build openbsd || netbsd || freebsd
// +build openbsd netbsd freebsd

package service

import (
	"context"
)

func ChooseBackend(ctx context.Context) (be Backend, err error) {
	for _, c := range choosers {
		if ok := c.Choose(ctx); ok {
			if be, ok = c.(Backend); ok {
				return
			}
		}
	}
	return
}

var (
	choosers = []Chooser{
		&sysvInitD{},
		&upstartD{},
	}
)
