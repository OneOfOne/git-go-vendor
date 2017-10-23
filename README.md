# git-go-vendor [![GoDoc](https://godoc.org/github.com/OneOfOne/git-go-vendor?status.svg)](https://godoc.org/github.com/OneOfOne/git-go-vendor)

A "super" simple git sub command to use vendor go packages using git submodule (which is automatically supported by go get).

## Install

1. `go get github.com/OneOfOne/git-go-vendor`
2. Make sure `$GOPATH/bin` is in your `$PATH`.
3. `git go-vendor -h`

## Usage

```
➤ git go-vendor -h
NAME:
   git-go-vendor - A new cli application

USAGE:
   git-go-vendor [global options] command [command options] [arguments...]

VERSION:
   v0.1

COMMANDS:
     list, ls    list all current directly vendored packages
     add, a      adds or replaces a vendor package {git-repo}[@branch|tag|hash] [alias]
     update, up  updates a vendored package or all of them if non is specified
     remove, rm  removes the vendor package
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbose, -v                  verbose output (default: false)
   --dry-run, -n                  dry-run, don't actually execute any git commands (default: false)
   --git-path value, --git value  path to the git executable (default: "git")
   --help, -h                     show help (default: false)
   --version, -V                  print the version (default: false)
```

## Example

```
━➤ git go-vendor a github.com/OneOfOne/xxhash@449a3a6b
* Added vendor/github.com/OneOfOne/xxhash @ v1.2-14-g449a3a6bec

━➤ git go-vendor ls
* vendor/github.com/OneOfOne/xxhash @ v1.2-14-g449a3a6bec

━➤ git go-vendor rm github.com/OneOfOne/xxhash
* Removed vendor/github.com/OneOfOne/xxhash

━➤ git go-vendor a github.com/OneOfOne/xxhash@449a3a6b github.com/OneOfOne/xxh
* Added github.com/OneOfOne/xxhash @ v1.2-14-g449a3a6bec → vendor/github.com/OneOfOne/xxh
━➤ git go-vendor ls
* github.com/OneOfOne/xxhash @ v1.2-14-g449a3a6bec → vendor/github.com/OneOfOne/xxh

━➤ git commit -a -m 'xxh vendoring'

```
## FAQ

### Why?

* Everything else is too complicated and depends on having the vendoring tool installed on the client.
* Managing extra config files overcomplicates vendoring.
* 90% of my use cases depends on using a git repo.

## TODO

* Look into supporting non-git repos.

## License

This project is released under the [MIT](https://opensource.org/licenses/MIT) license.
