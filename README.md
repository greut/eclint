# eclint - EditorConfig linter ★

An faster alternative to the [JavaScript _eclint_](https://github.com/jedmao/eclint) written in Go.

## Installation

- [Archlinux](https://aur.archlinux.org/packages/eclint/)
- [Docker](https://hub.docker.com/r/greut/eclint)
- [GitHub action](https://github.com/greut/eclint-action/)
- [Manual installs](https://gitlab.com/greut/eclint/-/releases)

## Usage

```
$ go install gitlab.com/greut/eclint/cmd/eclint

$ eclint -version
```

Excluding some files using the EditorConfig matcher

```
$ eclint -exclude "testdata/**/*"
```

## Features

- `charset`
- `end_of_line`
- `indent_size`
- `indent_style`
- `insert_final_newline`
- `max_line_length` (when using tabs, specify the `tab_width` or `indent_size`)
    - by default, UTF-8 charset is assumed and multi-byte characters should be
    counted as one. However, combining characters won't.
- `trim_trailing_whitespace`
- [domain-specific properties][dsl]
    - `line_comment`
    - `block_comment_start`, `block_comment`, `block_comment_end`
- miminal magic bytes detection (currently for PDF)

### More

- when not path is given, it searches for files via `git ls-files`
- `-exclude` to filter out some files
- unset / alter properties via the `eclint_` prefix
- [Docker images](https://hub.docker.com/r/greut/eclint) (also on GitHub and GitLab registries)
- colored output (use `-color`: `never` to disable and `always` to skip detection)
- `-summary` mode showing only the number of errors per file
- only the X first errors are shown (use `-show_all_errors` to disable)
- binary file detection (however quite basic)
- `-fix` to modify the file in place rather than showing the errors (currently only basic `unix2dos`, `dox2unix` is supported)

## Missing features

- `max_line_length` counting UTF-16 and UTF-32 characters
- more tests
- ability to fix: `insert_final_newline`, `indent_style`, `trim_trailing_whitespace`
- etc.

## Benchmarks

**NB** benchmarks matter at feature parity (which is also hard to measure).

The contenders are the following.

- [editorconfig-checker](https://github.com/editorconfig-checker/editorconfig-checker), also in Go.
- [eclint](https://github.com/jedmao/eclint), in Node.

The methodology is to run the linter against some big repositories `time eclint -show_all_errors`.

| Repository    | `editorconfig-checker` | `jedmao/eclint` | `greut/eclint` |
|---------------|------------------------|-----------------|----------------|
| [Roslyn][]    | 37s                    | 1m5s            | **4s**         |
| [SaltStack][] | 7s                     | 1m9s            | **<1s**        |

[Roslyn]: https://github.com/dotnet/roslyn
[SaltStack]: https://github.com/saltstack/salt

### Profiling

Two options: `-cpuprofile <file>` and `-memprofile <file>`, will produce the appropriate _pprof_ files.

## Libraries and tools

- [aurora](https://github.com/logrusorgru/aurora), colored output
- [chardet](https://github.com/gogs/chardet), charset detection
- [editorconfig-core-go](https://github.com/editorconfig/editorconfig-core-go), `.editorconfig` parsing
- [go-colorable](https://github.com/mattn/go-colorable), colored output on Windows (too soon)
- [go-mod-outdated](https://github.com/psampaz/go-mod-outdated)
- [golangci-lint](https://github.com/golangci/golangci-lint), Go linters
- [goreleaser](https://goreleaser.com/)
- [klogr](https://github.com/kubernetes/klog/tree/master/klogr)
- [nancy](https://github.com/sonatype-nexus-community/nancy)

[dsl]: https://github.com/editorconfig/editorconfig/wiki/EditorConfig-Properties#ideas-for-domain-specific-properties
