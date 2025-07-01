# 51Degrees IP Intelligence API

![51Degrees](https://51degrees.com/img/logo.png?utm_source=github&utm_medium=repository&utm_campaign=c_open_source&utm_content=readme_main "Data rewards the curious") **IP Intelligence in GO**

## Introduction

This repository contains a Go implementation of the IP Intelligence engine.
This module allows you to call functions for determining to set up with an IP Intelligence engine and begin using it to process IP addresses.

## Module Structure

**Module**: `ip-intelligence-go` at path `github.com/51Degrees/ip-intelligence-go`

This Go version contains the following packages:
- `ipi_interop` - a lower level API wrapping the C IP Intelligence library
- `ipi_onpremise` - a higher level Engine API providing IP Intelligence library and file update functions

## System Requirements

### Supported Platforms and Architectures
- Linux 32/64 bit, Intel and ARM processor
- MacOS 64 bit, Intel and ARM processor
- Windows 64bit, Intel

### Go Version
- Go 1.19

### Required Software
- Powershell Core (7 or above)
- A C compiler that support C11 or above (Gcc on Linux, Clang on MacOS and MinGW-x64 on Windows)
- libatomic - which usually come with default Gcc, Clang installation

## Setup Instructions

### Data File Setup

To process IP addresses, you will need to use the 51Degrees data file.
See [ip-intelligence-data](https://github.com/51Degrees/ip-intelligence-data) repo for the instructions on how to get a Lite data file or [contact us](https://51degrees.com/contact-us) to get an Enterprise data file.

### Windows Configuration

If you are on Windows, make sure that:
- The path to the `MinGW-x64` `bin` folder is included in the `PATH`. By default, the path should be `C:\msys64\ucrt64\bin`
- Go environment variable `CGO_ENABLED` is set to `1`

```
bash go env -w CGO_ENABLED=1
```


## Build and Usage

### Installation

Import the package as usual and it will get build automatically:


```
go import "github.com/51Degrees/ip-intelligence-go/ipi_onpremise"
```


### Vendored C Library

Implementation wraps an amalgamated C library from [ip-intelligence-cxx](https://github.com/51Degrees/ip-intelligence-cxx) repo

## Development

### Testing

Unit tests can be run with `go test ./...` from the root dir.

### API Documentation

To view APIs and their descriptions, users can use `go doc` in the package directory.
- First navigate to `ipi_interop` or `ipi_onpremise` dir.
- Then run the below to display all APIs, structures and their descriptions.


```
bash go doc -all
```

### Examples

Examples are included in a separate repository at [ip-intelligence-go-examples](https://github.com/51Degrees/ip-intelligence-go-examples).
Users can follow its README.md for details and comments in each examples for how to run them.
