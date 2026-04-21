package xls

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// WriteCSV writes UTF-16LE CSV with a BOM and a leading sep= line.
// Cell text must be UTF-8; encodingFrom is reserved (only UTF-8 is supported today).
func WriteCSV(w io.Writer, tab Table, delimiter, enclosure string, encodingFrom string, detectHead bool) error {
	if encodingFrom != "" && !strings.EqualFold(encodingFrom, "UTF-8") {
		return fmt.Errorf("xls: unsupported encodingFrom %q (only UTF-8)", encodingFrom)
	}
	if delimiter == "" {
		delimiter = ","
	}
	if enclosure == "" {
		enclosure = `"`
	}
	delimR, encR, err := csvRunes(delimiter, enclosure)
	if err != nil {
		return err
	}

	tab = normalizeTable(tab)

	var body strings.Builder
	_, _ = body.WriteString("sep=" + delimiter + "\n")

	if detectHead && len(tab.Rows) > 0 && len(tab.Columns) > 0 {
		writeCSVLine(&body, tab.Columns, delimR, encR)
		_, _ = body.WriteString("\n")
	}
	for _, row := range tab.Rows {
		rec := make([]string, len(tab.Columns))
		for i := range tab.Columns {
			if i < len(row) {
				rec[i] = row[i]
			}
		}
		writeCSVLine(&body, rec, delimR, encR)
		_, _ = body.WriteString("\n")
	}

	utf16le, err := utf8ToUTF16LE([]byte(body.String()))
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte{0xff, 0xfe}); err != nil {
		return err
	}
	_, err = w.Write(utf16le)
	return err
}

func csvRunes(delimiter, enclosure string) (delim rune, enc rune, err error) {
	drs := []rune(delimiter)
	if len(drs) != 1 {
		return 0, 0, fmt.Errorf("xls: delimiter must be one rune, got %q", delimiter)
	}
	ers := []rune(enclosure)
	if len(ers) != 1 {
		return 0, 0, fmt.Errorf("xls: enclosure must be one rune, got %q", enclosure)
	}
	return drs[0], ers[0], nil
}

func writeCSVLine(w *strings.Builder, fields []string, delim, enc rune) {
	for i, field := range fields {
		if i > 0 {
			_, _ = w.WriteRune(delim)
		}
		writeCSVField(w, field, delim, enc)
	}
}

func writeCSVField(w *strings.Builder, field string, delim, enc rune) {
	needsQuote := strings.ContainsRune(field, delim) ||
		strings.ContainsRune(field, enc) ||
		strings.ContainsRune(field, '\n') ||
		strings.ContainsRune(field, '\r')
	if !needsQuote {
		_, _ = w.WriteString(field)
		return
	}
	_, _ = w.WriteRune(enc)
	for _, r := range field {
		if r == enc {
			_, _ = w.WriteRune(enc)
		}
		_, _ = w.WriteRune(r)
	}
	_, _ = w.WriteRune(enc)
}

func utf8ToUTF16LE(b []byte) ([]byte, error) {
	if !utf8.Valid(b) {
		return nil, fmt.Errorf("xls: invalid utf-8 in csv body")
	}
	s := string(b)
	runes := []rune(s)
	u16 := utf16.Encode(runes)
	out := make([]byte, len(u16)*2)
	for i, v := range u16 {
		out[i*2] = byte(v)
		out[i*2+1] = byte(v >> 8)
	}
	return out, nil
}

// TrimTrailingFinalNewline removes one trailing UTF-16LE newline (after BOM) from b if present.
func TrimTrailingFinalNewline(b []byte) []byte {
	if len(b) < 2 || b[0] != 0xff || b[1] != 0xfe {
		return b
	}
	if len(b) >= 4 && b[len(b)-2] == 0x0A && b[len(b)-1] == 0x00 {
		return b[:len(b)-2]
	}
	return b
}
