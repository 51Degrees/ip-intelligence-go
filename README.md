# 51Degrees Ip Intelligence API

![51Degrees](https://51degrees.com/img/logo.png?utm_source=github&utm_medium=repository&utm_campaign=c_open_source&utm_content=readme_main "Data rewards the curious") **IP Intelligence in GO**

## Introduction

This repository contains Go Lite implementation of IP Intelligence engine.
(REPLACE)!!!! This provides a light set of device detection features that allow Go users to try and evaluate the capability of 51Degrees device detection solution in Go.

## Supported platforms and architectures
Go Lite implementation is currently supporting the following platforms and architectures:
- Linux 32/64 bit, Intel and ARM processor
- MacOS 64 bit, Intel and ARM processor
- Windows 64bit, Intel

Go version:
- 1.19

Compiles with go 1.19 and higher.  The minimum version of 1.19 is dictated by the fact that this library is used
by the [prebid-server](https://github.com/prebid/prebid-server/) module and it has a minimum version requirement of go 1.19.

## Pre-requisites

(REPLACE)!!!! ### Data File


### Software

In order to build use ip-intelligence-go the following are required:
- Powershell Core (7 or above)
- A C compiler that support C11 or above (Gcc on Linux, Clang on MacOS and MinGW-x64 on Windows)
- libatomic - which usually come with default Gcc, Clang installation

### Windows

If you are on Windows, make sure that:
- The path to the `MinGW-x64` `bin` folder is included in the `PATH`. By default, the path should be `C:\msys64\ucrt64\bin`
- Go environment variable `CGO_ENABLED` is set to `1`
```
go env -w CGO_ENABLED=1
```

(REPLACE)!!!! ## Module and Packages


## Build and Usage

### Build steps for all platforms

Import the package as usual and it will get build automatically:

```
import "github.com/51Degrees/ip-intelligence-go/ipi_onpremise"
```

### Vendored C Library
An amalgamation is an alternative way to distribute a library's source code using only a few files (as low as one or two).
This go module depends on and ships an amalgamation of the device-detection C-library [ip-intelligence-cxx](https://github.com/51degrees/ip-intelligence-cxx) repository, which is then built automatically by CGo during `go build`.

The amalgamation is produced automatically and regularly by the [Nightly Package Update](https://github.com/51Degrees/ip-intelligence-go/actions/workflows/nightly-package-update.yml) CI workflow from the ip-intelligence C-library source code.

## Test

Unit tests can be run with `go test ./...` from the root dir.

(REPLACE)!!!! ## APIs

**NOTE**: Not all APIs are available. Ones that are not available include string `TODO: To be implemented` in their descriptions.


## Examples

Examples are included in a separate repository at [ip-intelligence-examples-go](https://github.com/51degrees/ip-intelligence-examples-go). Users can follow its README.md for details and comments in each examples for how to run them.