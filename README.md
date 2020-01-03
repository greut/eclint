# eclint - EditorConfig linter ★

An alternative to the [JavaScript _eclint_](https://github.com/jedmao/eclint) written in Go.

**Work in progress**

## Usage

```
$ go install gitlab.com/greut/eclint

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
- [Docker images](https://hub.docker.com/r/greut/eclint)
- colored output _à la_ ripgrep (use `-no_colors` to disable, or `-force_colors` to skip detection)
- `-summary` mode showing only the number of errors per file
- only the X first errors are shown (use `-show_all_errors` to disable)
- binary file detection (however quite basic)

## Missing features

- basic `//nolint` [suffix](https://github.com/golangci/golangci-lint#nolint)
- `max_line_length` counting UTF-16 and UTF-32 characters
- more tests
- ability to fix
- etc.

## Benchmarks

**NB** benchmarks matter at feature parity (which is also hard to measure).

The contenders are the following.

- [editorconfig-checker](https://github.com/editorconfig-checker/editorconfig-checker), also in Go.
- [eclint](https://github.com/jedmao/eclint), in Node.

The methodology is to run the linter against some big repositories `time $(eclint >/dev/null)`.

| Repository | `editorconfig-checker` | `jedmao/eclint` | `greut/eclint` |
|------------|------------------------|-----------------|----------|
| [Roslyn](https://github.com/dotnet/roslyn) | 37s | 1m5s | **8s** |
| [SaltStack](https://github.com/saltstack/salt) | 7s | 1m9s |  **1s** |

## Libraries and tools

- [aurora](https://github.com/logrusorgru/aurora), colored output
- [chardet](https://github.com/gogs/chardet), charset detection
- [editorconfig-core-go](https://github.com/editorconfig/editorconfig-core-go), `.editorconfig` parsing
- [go-colorable](https://github.com/mattn/go-colorable), colored output on Windows (too soon)
- [golangci-lint](https://github.com/golangci/golangci-lint), Go linters
- [goreleaser](https://goreleaser.com/)
- [klogr](https://github.com/kubernetes/klog/tree/master/klogr)

[dsl]: https://github.com/editorconfig/editorconfig/wiki/EditorConfig-Properties#ideas-for-domain-specific-properties
