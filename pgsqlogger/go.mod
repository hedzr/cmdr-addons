module github.com/hedzr/cmdr-addons/pgsqlogger/v2

go 1.24.0

toolchain go1.24.5

// replace github.com/hedzr/cmdr-addons => ../

// replace github.com/hedzr/cmdr-addons/v2 => ../

// replace github.com/hedzr/cmdr/v2 => ../cmdr

// replace gopkg.in/hedzr/errors.v3 => ../../24/libs.errors

require (
	github.com/hedzr/logg v0.8.66
	github.com/lib/pq v1.10.9
)

require (
	github.com/hedzr/is v0.8.66 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/term v0.38.0 // indirect
	gopkg.in/hedzr/errors.v3 v3.3.5 // indirect
)
