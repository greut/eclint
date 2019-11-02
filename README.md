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

## Missing features

- `utf-8 bom`
- [domain-specific properties](https://github.com/editorconfig/editorconfig/wiki/EditorConfig-Properties#ideas-for-domain-specific-properties)
    - `line_comment`
    - `block_comment_start`, `block_comment`, `block_comment_end`
- more tests
- ability to fix
- ignoring `.git`, `vendor`, etc. (via `git ls-files`)
- docker image
- etc.

## Benchmarks

*TODO*

## Libraries

- [editorconfig-core-go](https://github.com/editorconfig/editorconfig-core-go)
- [klogr](https://github.com/kubernetes/klog/tree/master/klogr)
- [chardet](https://github.com/gogs/chardet)
