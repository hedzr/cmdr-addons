module github.com/hedzr/cmdr-addons/service/v2

go 1.24.0

toolchain go1.24.5

replace github.com/hedzr/cmdr-addons => ../

replace github.com/hedzr/cmdr-addons/v2 => ../

// replace github.com/hedzr/cmdr/v2 => ../cmdr

// replace gopkg.in/hedzr/errors.v3 => ../../24/libs.errors

require (
	github.com/hedzr/cmdr-addons/v2 v2.0.21
	github.com/hedzr/is v0.8.61
	github.com/hedzr/logg v0.8.61
	github.com/stretchr/testify v1.11.1
	golang.org/x/sys v0.37.0
	gopkg.in/hedzr/errors.v3 v3.3.5
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/term v0.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
