module github.com/hedzr/cmdr-addons/v2/examples

go 1.23.0

toolchain go1.23.3

replace github.com/hedzr/cmdr-addons/v2 => ../

replace github.com/hedzr/cmdr-addons/v2/service/v2 => ../service

require (
	github.com/hedzr/cmdr-addons/v2 v2.0.3
	github.com/hedzr/cmdr-addons/v2/service/v2 v2.0.3
	github.com/hedzr/cmdr-loaders v1.3.39
	github.com/hedzr/cmdr/v2 v2.1.39
	github.com/hedzr/is v0.8.39
	github.com/hedzr/store v1.3.39
	gopkg.in/hedzr/errors.v3 v3.3.5
)

require (
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hedzr/evendeep v1.3.39 // indirect
	github.com/hedzr/logg v0.8.39 // indirect
	github.com/hedzr/store/codecs/hcl v1.3.39 // indirect
	github.com/hedzr/store/codecs/hjson v1.3.39 // indirect
	github.com/hedzr/store/codecs/json v1.3.39 // indirect
	github.com/hedzr/store/codecs/nestext v1.3.39 // indirect
	github.com/hedzr/store/codecs/toml v1.3.39 // indirect
	github.com/hedzr/store/codecs/yaml v1.3.39 // indirect
	github.com/hedzr/store/providers/env v1.3.39 // indirect
	github.com/hedzr/store/providers/file v1.3.39 // indirect
	github.com/hjson/hjson-go/v4 v4.5.0 // indirect
	github.com/npillmayer/nestext v0.1.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
