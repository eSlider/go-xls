# go-xls

[![CI](https://github.com/eSlider/go-xls/actions/workflows/ci.yml/badge.svg)](https://github.com/eSlider/go-xls/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/eSlider/go-xls.svg)](https://pkg.go.dev/github.com/eSlider/go-xls)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)

Small Go library for **CSV**, legacy **.xls** (BIFF), and **.xlsx** export, ported from Mapbender’s [`ExportResponse.php`](https://github.com/mapbender/mapbender/blob/master/src/FOM/CoreBundle/Component/ExportResponse.php) (`FOM\CoreBundle\Component\ExportResponse`). It focuses on the same wire formats and HTTP headers used there: UTF‑16LE CSV with BOM and `sep=`, hand-written BIFF cells for `.xls`, and OpenXML `.xlsx` via [excelize](https://github.com/xuri/excelize).

## Install

```bash
go get github.com/eSlider/go-xls
```

## Usage

```go
package main

import (
	"net/http"
	"os"

	"github.com/eSlider/go-xls"
)

func main() {
	tab := xls.Table{
		Columns: []string{"name", "qty"},
		Rows: [][]string{
			{"apple", "3"},
			{"banana", "2"},
		},
	}

	// Legacy .xls (Mapbender genXLS-compatible layout on LE).
	bin, err := xls.GenXLS(tab, true)
	if err != nil {
		panic(err)
	}
	_ = os.WriteFile("export.xls", bin, 0o644)

	// UTF-16LE CSV with BOM + sep= line.
	csv, err := xls.GenCSV(tab, ",", `"`, "UTF-8", true)
	if err != nil {
		panic(err)
	}
	_ = os.WriteFile("export.csv", csv, 0o644)

	// .xlsx
	xlsx, err := xls.GenXLSX(tab, true)
	if err != nil {
		panic(err)
	}
	_ = os.WriteFile("export.xlsx", xlsx, 0o644)

	// HTTP attachment (Symfony-style headers from Mapbender).
	_ = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := xls.GenXLS(tab, true)
		xls.WriteAttachment(w, "export.xls", xls.ContentTypeXLS, body, true)
	})
}
```

## Behaviour notes

- **Table**: columns are explicit and stable (Go has no PHP ordered-map iteration). If `Columns` is empty, synthetic names `0`, `1`, … are derived from the widest row, matching numeric-key tables in Mapbender.
- **XLS**: string cells are encoded to **ISO‑8859‑1** (unmappable runes become `?`). Numbers use `strconv.ParseFloat` after `strings.TrimSpace`, similar to PHP `is_numeric` / `trim` usage.
- **CSV**: only **`UTF-8` → UTF‑16LE** is implemented for `encodingFrom`; other labels return an error until transcoding is extended.
- **XLSX**: uses **excelize**; numeric-looking cells are written as floats.

## Testing

```bash
go test -v -race ./...
```

## License

[MIT](LICENSE)
