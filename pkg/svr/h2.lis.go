// Copyright © 2020 Hedzr Yeh.

package svr

import (
	"net"
)

func getOnGetListener() net.Listener {
	// l := h2listener
	// h2listener = nil
	// return l
	return h2listener
}

var h2listener net.Listener
