// Package ole documents the binary lead-in of Microsoft Compound File Binary (CFB) files,
// also called OLE compound documents or “Structured Storage”. Excel 97–2003 often stores
// .xls workbooks inside this container; those files are not a raw linear BIFF byte stream.
//
// [github.com/eslider/go-xls/v2.ReadXLS] accepts only the minimal linear BIFF layout produced by
// [github.com/eslider/go-xls/v2.WriteXLS]. If the input begins with [HeaderPrefix], it returns
// [github.com/eslider/go-xls/v2.ErrOLEWorkbook] so callers can branch to an OLE-capable
// reader (e.g. full CFB + BIFF8) instead of mis-parsing container bytes as records.
//
// Specification context: ECMA-376 / MS-CFB describe the CFB format; the file header’s first
// eight bytes are the fixed “magic” pattern below (often shown as one 64-bit signature).
package ole

import "bytes"

// HeaderPrefix is the first four bytes of the CFB / OLE2 file header (little-endian as stored on disk).
// In hex: D0 CF 11 E0. The next four bytes of a complete compound file are usually A1 B1 1A E1;
// this package and [github.com/eslider/go-xls/v2.ReadXLS] only test this four-byte prefix to
// distinguish container .xls from a linear BIFF stream (which typically starts with BOF record 09 08).
//
// Do not confuse with a BIFF [github.com/eslider/go-xls/v2/biff.RecordBOF] (0x0809 on the wire as 09 08 LE).
var HeaderPrefix = [...]byte{0xD0, 0xCF, 0x11, 0xE0}

// HasHeaderPrefix reports whether b is at least len([HeaderPrefix]) bytes long and starts with
// that prefix. Short buffers return false (no panic); empty input returns false.
func HasHeaderPrefix(b []byte) bool {
	return len(b) >= len(HeaderPrefix) && bytes.HasPrefix(b, HeaderPrefix[:])
}
