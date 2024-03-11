# Go Version Selector (GVS)

GVS automatically determines and executes only the main Go version, but GVS does not manage tools by versions.

## Install

```
go install github.com/komem3/gvs@latest

nvs init

# Add PATH
# export PATH="$HOME/.gvs/bin:$PATH"

nvs use 1.22
go --version
```

## Version Determination

1. read `.go-version` in current path.
2. read `go` field of `go.mod` in current path.
3. go to the parent directory. Back to 1. If there are no more parents, Go to 4.
4. read global version file(`$HOME/.gvs/version`)

## Install Global Tool

If you want to install a tool in a global version instead of a local version,
you can use the global version with the following command.

```
gvs install golang.org/x/tools/cmd/goimports@latest
```

## Usage

```
Usage:
  gvs [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  download    Download specify version of Go
  help        Help about any command
  init        Initialize gvs
  install     install tools by global Go version
  run         Run command(go or gofmt)
  use         Select Go version
  versions    List version

Flags:
      --debug   output debug log
  -h, --help    help for gvs

Use "gvs [command] --help" for more information about a command.
```
