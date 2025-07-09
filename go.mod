module github.com/hedzr/cmdr-addons/v2

go 1.23.0

toolchain go1.23.3

// replace github.com/hedzr/cmdr/v2 => ../cmdr

// replace gopkg.in/hedzr/errors.v3 => ../../24/libs.errors

require github.com/hedzr/logg v0.8.39

require (
	github.com/hedzr/is v0.8.39 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	gopkg.in/hedzr/errors.v3 v3.3.5 // indirect
)
