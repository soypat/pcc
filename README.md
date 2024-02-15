# pcc
[![go.dev reference](https://pkg.go.dev/badge/github.com/soypat/pcc)](https://pkg.go.dev/github.com/soypat/pcc)
[![Go Report Card](https://goreportcard.com/badge/github.com/soypat/pcc)](https://goreportcard.com/report/github.com/soypat/pcc)
[![codecov](https://codecov.io/gh/soypat/pcc/branch/main/graph/badge.svg)](https://codecov.io/gh/soypat/pcc)
[![Go](https://github.com/soypat/pcc/actions/workflows/go.yml/badge.svg)](https://github.com/soypat/pcc/actions/workflows/go.yml)
[![sourcegraph](https://sourcegraph.com/github.com/soypat/pcc/-/badge.svg)](https://sourcegraph.com/github.com/soypat/pcc?badge)
[![License: BSD-3Clause](https://img.shields.io/badge/License-BSD-3.svg)](https://opensource.org/licenses/bsd-3-clause)

pcc solves several questions I've had over the years of working with industrial processes:

- I want to define logic on the Process Controller side and have it be discoverable by an operator.
- I want to be able to represent sequential and simultaneous processes and serialize them reliably and unambiguously.
    - I want a simple binary protocol for serializing process configuration.
    - Modbus registers for configuring process controllers is great, let's do something like that but reduced in scope and with extra register metadata in mind such as SI units.
- I want to reuse modules between projects without duplicating too much code.
- I want generic database interoperability with my process.
- I want constrained memory consumption to be able to use protocol on microcontrollers. TinyGo compatibility.

How to install package:
```sh
go mod download github.com/soypat/pcc@latest
```

This is a work in progress. pcc provides robust process control primitives. 

