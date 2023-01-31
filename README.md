# cmdr-addons: Addons for `cmdr`

![Go](https://github.com/hedzr/cmdr-addons/workflows/Go/badge.svg)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/hedzr/cmdr-addons.svg?label=release)](https://github.com/hedzr/cmdr-addons/releases)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fhedzr%2Fcmdr-addons.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fhedzr%2Fcmdr-addons?ref=badge_shield)

see also [`cmdr`](https://github.com/hedzr/cmdr).

> **NOTE**  
> The `cmdr-addons` version is following `cmdr`'s now.

## Prerequisites

### v1.11.6 and newer

golang 1.17+ required

> The constraints are from updating via go modules.

### v1.9.8-p3 and newer

golang 1.15+ required.

> **Causes**:  
> 1. golang.org/x/net/http2 used errors.Is()
> 2. golang.org/x/net/http2 used os.ErrDeadlineExceeded

Updates:
1. removed iris/v12 [`import "github.com/hedzr/cmdr-addons v1.9.8-p3"`]
2. seems ci not good for go1.14



### v1.9.7 and older

golang 1.13+ required.

> **Causes**:  
>   github.com/kataras/iris/v12@v12.1.8/core/errgroup/errgroup.go:109:9: undefined: errors.Unwrap
>
> **Workaround**:  
>   Avoid using `iris` codes in `svr` templates.



## Includes:

### Plugins

#### `dex`

- new version of `daemon` plugin: `dex`

- sample app:
  For examples, see also: [the example app: service](https://github.com/hedzr/cmdr-examples/tree/master/examples/service)

#### `svr`

The wrapped http2 server with multiple 3rd multiplexers (echo, gin, ...).

sample app:

[the example app: service](https://github.com/hedzr/cmdr-examples/tree/master/examples/service)


#### `trace`
- `trace`: adds `--trace` to your root command

```go
TODO
```


### Others

- `svr`: template codes for http/2 server (mux)
- `vxconf`: helpers



## Thanks to JODL

[JODL (JetBrains OpenSource Development License)](https://www.jetbrains.com/community/opensource/) is good:

[![goland](https://gist.githubusercontent.com/hedzr/447849cb44138885e75fe46f1e35b4a0/raw/ca8ac2694906f5650d585263dbabfda52072f707/logo-goland.svg)](https://www.jetbrains.com/?from=hedzr/cmdr-addons)
[![jetbrains](https://gist.githubusercontent.com/hedzr/447849cb44138885e75fe46f1e35b4a0/raw/bedfe6923510405ade4c034c5c5085487532dee4/jetbrains-variant-4.svg)](https://www.jetbrains.com/?from=hedzr/cmdr-addons)



## License

MIT






[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fhedzr%2Fcmdr-addons.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fhedzr%2Fcmdr-addons?ref=badge_large)