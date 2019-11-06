# eclint - EditorConfig linter ★

An alternative to the [JavaScript _eclint_](https://github.com/jedmao/eclint) written in Go.

**Work in progress**

## Usage

```
$ go install gitlab.com/greut/eclint

$ eclint -version
```

## Features

- `charset`
- `end_of_line`
- `indent_size`
- `indent_style`
- `insert_final_newline`
- `trim_trailing_whitespace`
- when not path is given, it searches for files via `git ls-files`

## Missing features

- doing checks on `rune` rather than `byte`
- [domain-specific properties](https://github.com/editorconfig/editorconfig/wiki/EditorConfig-Properties#ideas-for-domain-specific-properties)
    - `line_comment`
    - `block_comment_start`, `block_comment`, `block_comment_end`
- more tests
- colored output _à la_ ripgrep
- ability to fix
- docker image
- etc.

## Benchmarks

**NB** benchmarks matter at feature parity (which is also hard to measure).

The contenders are the following.

- [editorconfig-checker](https://github.com/editorconfig-checker/editorconfig-checker), also in Go.

The methodology is to run the linter against some big repositories such as:

- Rosylin
- SaltStack

## Libraries

- [editorconfig-core-go](https://github.com/editorconfig/editorconfig-core-go)
- [klogr](https://github.com/kubernetes/klog/tree/master/klogr)
- [chardet](https://github.com/gogs/chardet)
