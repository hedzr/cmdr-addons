# Addons for cmdr/v2

![Go](https://github.com/hedzr/cmdr-addons/workflows/release-build/badge.svg)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/hedzr/cmdr-addons)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/hedzr/cmdr-addons.svg?label=release)](https://github.com/hedzr/cmdr-addons/releases)

This is an addons repo especially for [cmdr/v2](https://github.com/hedzr/cmdr).

The typical app is [cmdr-addons/examples/myservice](https://github.com/hedzr/cmdr-addons/blob/master/examples/myservice).

![image-20241111141228632](https://cdn.jsdelivr.net/gh/hzimg/blog-pics@master/upgit/2024/11/20241111_1731305562.png)

A tiny app using `cmdr/v2` and `cmdr-addons` is:

```go
//
```

See also:

- [cmdr/v2](https://github.com/hedzr/cmdr)

## Addons

### `Service`

`github.com/hedzr/cmdr-addons/service/v2`, as its name hints, is a cross-platform service wrapper for macOS, Linux and Windows.

### `pgsqlogger`

`github.com/hedzr/cmdr-addons/pgsqlogger/v2`, is a pluggable `Writer` for `hedzr/logg/slog`, which can copy the logging lines into postgresql.

## History

See full list in [CHANGELOG](https://github.com/hedzr/cmdr-addons/blob/master/CHANGELOG)

## License

Apache 2.0
