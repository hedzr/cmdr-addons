package service

import (
	"strings"
)

type Entity interface {
	Name() string        // app-name, service-name, executable-name
	Desc() string        //
	ScreenName() string  //
	ServiceName() string // used by service manager, typically is app-name.service (for systemd)
	ExecutablePath() string
}

type entityS struct {
	name        string `json:"name,omitempty"`
	description string `json:"description,omitempty"`
	screenName  string `json:"screenName,omitempty"`
	executable  string `json:"executable,omitempty"`
}

func (e *entityS) String() string {
	return e.name
}

func (e *entityS) Name() string       { return e.name }
func (e *entityS) ScreenName() string { return e.screenName }
func (e *entityS) ServiceName() string {
	n := e.screenName
	if n == "" {
		n = e.name
	}
	if !strings.HasSuffix(n, ".service") {
		n += ".service"
	}
	return n
}

func (e *entityS) Description() string { return e.description }
func (e *entityS) Executable() string  { return e.name }

func (e *entityS) Port() int { return -1 }

//

//

//
