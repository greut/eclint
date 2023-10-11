# eclint - EditorConfig linter ★

A faster alternative to the [JavaScript _eclint_](https://github.com/jedmao/eclint) written in Go.

Tarballs are signed (`.minisig`) using the following public key:

    RWRP3/Z4+t+iZk1QU6zufn6vSDlvd76FLWhGCkt5kE7YqW3mOtSh7FvE

Which can be verified using [minisig](https://jedisct1.github.io/minisign/) or [signify](https://github.com/aperezdc/signify).

## Installation

- [Archlinux](https://aur.archlinux.org/packages/eclint/)
- [Docker](https://hub.docker.com/r/greut/eclint) ([Quay.io](https://quay.io/repository/greut/eclint))
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
- minimal magic bytes detection (currently for PDF)

### More

- when no path is given, it searches for files via `git ls-files`
- `-exclude` to filter out some files
- unset / alter properties via the `eclint_` prefix
- [Docker images](https://hub.docker.com/r/greut/eclint) (also on Quay.io, GitHub and GitLab registries)
- colored output (use `-color`: `never` to disable and `always` to skip detection)
- `-summary` mode showing only the number of errors per file
- only the first X errors are shown (use `-show_all_errors` to disable)
- binary file detection (however quite basic)
- `-fix` to modify files in place rather than showing the errors currently:
    - only basic `unix2dos`, `dos2unix`
    - space to tab and tab to space conversion
    - trailing whitespaces

## Missing features

- `max_line_length` counting UTF-32 characters
- more tests
- etc.

## Thanks for their contributions

- [Viktor Szépe](https://github.com/szepeviktor)
- [Takuya Fukuju](https://github.com/chalkygames123)
- [Nicolas Mohr](https://gitlab.com/nicmr)
- [Zadkiel Aharonian](https://gitlab.com/zadkiel_aharonian_c4)

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
