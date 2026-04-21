# go-xls

[![CI](https://github.com/eSlider/go-xls/actions/workflows/ci.yml/badge.svg)](https://github.com/eSlider/go-xls/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/eSlider/go-xls.svg)](https://pkg.go.dev/github.com/eSlider/go-xls)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)
[![Latest Release](https://img.shields.io/github/v/tag/eSlider/go-xls?sort=semver&label=release)](https://github.com/eSlider/go-xls/releases)

Small, **`io.Writer` / `io.Reader`–first** helpers for tabular data: UTF‑16LE CSV (BOM + `sep=`), legacy binary **.xls** (linear BIFF), **.xlsx** via [excelize](https://github.com/xuri/excelize), GitHub-style **markdown pipe tables**, and optional HTTP attachment headers.

## Install

```bash
go get github.com/eSlider/go-xls@v1.0.0
```

The string **`xls.Version`** matches this release (`1.0.0`). Git tags use a `v` prefix (`v1.0.0`); keep the const and tag in sync when publishing.

## Write legacy `.xls`

```go
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

Linear BIFF only (BOF `0x809`, string `0x204`, number `0x203`, EOF `0x0A`). OLE compound workbooks (`D0 CF 11 E0 …`) return `xls.ErrOLEWorkbook` — use another reader or convert to `.xlsx`.

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

## Write `.xlsx`

```go
var out bytes.Buffer
if err := xls.WriteXLSX(&out, tab, true); err != nil {
	log.Fatal(err)
}
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
