# eclint - EditorConfig linter ★

An alternative to the JavaScript _eclint_ written in Go.

## Usage

```
$ go install gitlab.com/greut/eclint

$ eclint -version
```

## Features

- `charset`
- `end_of_line`
- `indent_style`
- `indent_size`
- `insert_final_newline`
- `trim_trailing_whitespace`

## Missing features

- [domain-specific properties](https://github.com/editorconfig/editorconfig/wiki/EditorConfig-Properties#ideas-for-domain-specific-properties)
    - `line_comment`
    - `block_comment_start`, `block_comment`, `block_comment_end`
- more tests
- ability to fix
- ignoring `.git`, `vendor`, etc.
- docker image
- etc.

## Benchmarks

*TODO*

## Libraries

- [editorconfig-core-go](https://github.com/editorconfig/editorconfig-core-go)
