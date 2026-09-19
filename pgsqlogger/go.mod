module github.com/hedzr/cmdr-addons/pgsqlogger/v2

go 1.26.0

// replace github.com/hedzr/cmdr-addons => ../

// replace github.com/hedzr/cmdr-addons/v2 => ../

// replace github.com/hedzr/cmdr/v2 => ../cmdr

// replace gopkg.in/hedzr/errors.v3 => ../../24/libs.errors

require (
	github.com/hedzr/logg v0.9.9
	github.com/lib/pq v1.12.3
)

require (
	github.com/hedzr/is v0.9.9 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/term v0.46.0 // indirect
	gopkg.in/hedzr/errors.v3 v3.3.5 // indirect
)
