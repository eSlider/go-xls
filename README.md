# go-xls

[![CI](https://github.com/eSlider/go-xls/actions/workflows/ci.yml/badge.svg)](https://github.com/eSlider/go-xls/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/eslider/go-xls/v2.svg)](https://pkg.go.dev/github.com/eslider/go-xls/v2)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)
[![Latest Release](https://img.shields.io/github/v/tag/eSlider/go-xls?sort=semver&label=release)](https://github.com/eSlider/go-xls/releases)

**[pkg.go.dev/github.com/eslider/go-xls](https://pkg.go.dev/github.com/eslider/go-xls)** (module index) · **[pkg.go.dev/…/v2](https://pkg.go.dev/github.com/eslider/go-xls/v2)** (this major version’s `xls` package)

If those pages show *Not Found* right after a release, wait a few minutes for pkg.go.dev to index, or run `go doc github.com/eslider/go-xls/v2`. The module resolves from the proxy (`go get github.com/eslider/go-xls/v2@v2.1.0`).

Small, **`io.Writer` / `io.Reader`–first** helpers for tabular data: UTF‑16LE CSV (BOM + `sep=`), legacy binary **.xls** (linear BIFF), GitHub-style **markdown pipe tables**, and optional HTTP attachment headers.

## Install

```bash
go get github.com/eslider/go-xls/v2@v2.1.0
```

**Module path:** `github.com/eslider/go-xls/v2` (Go semantic import versioning for v2+).

## Write legacy `.xls`

```go
import "github.com/eslider/go-xls/v2"

tab := xls.Table{
	Columns: []string{"name", "qty"},
	Rows:    [][]string{{"apple", "3"}},
}
var buf bytes.Buffer
if err := xls.WriteXLS(&buf, tab, true); err != nil {
	log.Fatal(err)
}
```

## Read legacy `.xls`

Linear BIFF only (see [`biff` record opcodes](https://pkg.go.dev/github.com/eslider/go-xls/v2/biff#pkg-constants): `biff.RecordBOF`, `biff.RecordString`, `biff.RecordNumber`, `biff.RecordEOF`). OLE compound workbooks (magic `D0 CF 11 E0 …`) return `xls.ErrOLEWorkbook` — use another reader for those files; the signature is [`ole.HeaderPrefix`](https://pkg.go.dev/github.com/eslider/go-xls/v2/ole#pkg-variables).

```go
tab, err := xls.ReadXLS(bytes.NewReader(buf.Bytes()), true)
if err != nil {
	log.Fatal(err)
}
maps, err := xls.ReadXLSToMaps(bytes.NewReader(buf.Bytes()))
_ = maps
```

## Write UTF‑16LE CSV

```go
var out bytes.Buffer
err := xls.WriteCSV(&out, tab, ",", `"`, "UTF-8", true)
```

## Markdown pipe tables

```go
var md bytes.Buffer
if err := xls.WriteMarkdownTable(&md, tab); err != nil {
	log.Fatal(err)
}
parsed, err := xls.ReadMarkdownTable(bytes.NewReader(md.Bytes()))
_, _, err = xls.ReadMarkdownTableDetailed(bytes.NewReader(md.Bytes()))
_ = parsed
```

Alignment round-trip: `WriteMarkdownTableWith` + `ReadMarkdownTableDetailed` + `MarkdownMarshalOpts{Align: …}`.

## HTTP attachment

(`bytes` and `net/http` imports omitted.)

```go
body := bytes.NewReader(buf.Bytes())
if err := xls.WriteAttachment(w, "export.xls", xls.ContentTypeXLS, body, int64(buf.Len()), true); err != nil {
	log.Fatal(err)
}
```

Use `size < 0` to omit `Content-Length` (chunked transfer).

## Behaviour notes

- **`Table`**: if `Columns` is empty, names `"0"`, `"1"`, … are derived from the widest row.
- **`.xls` strings**: ISO‑8859‑1 bytes; unmappable runes become `?`. Numeric-looking cells use `strconv.ParseFloat` on trimmed text.
- **CSV**: only UTF‑8 source → UTF‑16LE is implemented for `encodingFrom` today.
- **Markdown**: `|` and `\` in cells are escaped; pipe tables follow common GitHub-style divider rules.

## Testing

```bash
go test -v -race ./...
```

## License

[MIT](LICENSE)
