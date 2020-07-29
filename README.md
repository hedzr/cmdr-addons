# Addons for `cmdr`

![Go](https://github.com/hedzr/cmdr-addons/workflows/Go/badge.svg)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/hedzr/cmdr-addons.svg?label=release)](https://github.com/hedzr/cmdr-addons/releases)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fhedzr%2Fcmdr-addons.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fhedzr%2Fcmdr-addons?ref=badge_shield)

see also [`cmdr`](https://github.com/hedzr/cmdr).

## Prerequisites

golang 1.13+ ONLY!

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