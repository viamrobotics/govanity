# govanity

For nice vanity urls for golang and private documentation.

Example @ http://go.viam.com/

## Setup

In general, you must set `GO_MODULES` to the modules you wish to provide vanity URLs for. For documentation of
private modules, you do not have to set a vanity pairing. For example:

`export GO_MODULES=vanity.domain.com/private-repo=github.com/org/private-repo,github.com/org/no-vanity-private-repro`

## Running

### With just vanity

`go run cmd/server/main.go`

### With docs

In addition to the `GO_MODULES`, you must point `govanity` to a path containing a `go.mod` in order to know where to find source code. It's expected that the module has already been `go mod download`ed.

`go run cmd/server/main.go --enable-docs path/to/module`

## Bootstrapping

If you need to run a server that will not have access to the module contents in advance,
you should bootstrap it. This will get the latest versions of the modules.

`make docs_bootstrap`

This will generate a `go.mod` file at `./doc/mod/go.mod` with the latest versions of the modules that you can use to run `govanity`.

## Static Doc Generation

`make docs_static`
