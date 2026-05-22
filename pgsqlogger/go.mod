module github.com/hedzr/cmdr-addons/pgsqlogger/v2

go 1.25.0

// replace github.com/hedzr/cmdr-addons => ../

// replace github.com/hedzr/cmdr-addons/v2 => ../

// replace github.com/hedzr/cmdr/v2 => ../cmdr

// replace gopkg.in/hedzr/errors.v3 => ../../24/libs.errors

require (
	github.com/hedzr/logg v0.9.0
	github.com/lib/pq v1.12.3
)

require (
	github.com/hedzr/is v0.9.1 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/term v0.41.0 // indirect
	gopkg.in/hedzr/errors.v3 v3.3.5 // indirect
)
