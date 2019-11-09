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
- `trim_trailing_whitespace`
- [domain-specific properties](https://github.com/editorconfig/editorconfig/wiki/EditorConfig-Properties#ideas-for-domain-specific-properties)
    - `line_comment`
    - `block_comment_start`, `block_comment`, `block_comment_end`
- when not path is given, it searches for files via `git ls-files`
- `-exclude` to filter out some files
- docker image

## Missing features

- basic `//nolint` [suffix](https://github.com/golangci/golangci-lint#nolint)
- doing checks on `rune` rather than `byte`
- more tests
- colored output _à la_ ripgrep
- ability to fix
- etc.

## Benchmarks

**NB** benchmarks matter at feature parity (which is also hard to measure).

The contenders are the following.

- [editorconfig-checker](https://github.com/editorconfig-checker/editorconfig-checker), also in Go.

The methodology is to run the linter against some big repositories `time $(eclint >/dev/null)`.

| Repository | `editorconfig-checker` | `eclint` |
|------------|------------------------|----------|
| [Roslyn](https://github.com/dotnet/roslyn) | 37s | 14s |
| [SaltStack](https://github.com/saltstack/salt) | 7s | 0.7s |

## Libraries and tools

- [chardet](https://github.com/gogs/chardet)
- [editorconfig-core-go](https://github.com/editorconfig/editorconfig-core-go)
- [golangci-lint](https://github.com/golangci/golangci-lint)
- [goreleaser](https://goreleaser.com/)
- [klogr](https://github.com/kubernetes/klog/tree/master/klogr)
