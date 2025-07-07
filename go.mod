module github.com/hedzr/cmdr-addons/v2

go 1.23.0

toolchain go1.23.3

// replace github.com/hedzr/cmdr/v2 => ../cmdr

// replace gopkg.in/hedzr/errors.v3 => ../../24/libs.errors

require (
	github.com/hedzr/cmdr/v2 v2.1.35
	github.com/hedzr/is v0.8.35
	github.com/hedzr/logg v0.8.35
	github.com/stretchr/testify v1.10.0
	golang.org/x/sys v0.33.0
	gopkg.in/hedzr/errors.v3 v3.3.5
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/hedzr/evendeep v1.3.35 // indirect
	github.com/hedzr/store v1.3.35 // indirect
	github.com/hedzr/store/codecs/json v1.3.35 // indirect
	github.com/hedzr/store/providers/file v1.3.35 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
