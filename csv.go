package xls

import (
	"encoding/binary"
	"encoding/csv"
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

// ReadCSV parses CSV produced by [WriteCSV]: a UTF-16LE BOM, a leading sep= line, then RFC 4180–style records.
// delimiter and enclosure must match the writer (only ASCII double quote `"` is supported for enclosure).
// When firstRowAsHeader is true, the first record becomes [Table.Columns] and the rest [Table.Rows];
// when false, every record is a row and column names are synthesized like [normalizeTable].
func ReadCSV(r io.Reader, delimiter, enclosure string, firstRowAsHeader bool) (Table, error) {
	if delimiter == "" {
		delimiter = ","
	}
	if enclosure == "" {
		enclosure = `"`
	}
	if enclosure != `"` {
		return Table{}, fmt.Errorf("xls: ReadCSV only supports enclosure %q", `"`)
	}
	delimR, _, err := csvRunes(delimiter, enclosure)
	if err != nil {
		return Table{}, err
	}

	raw, err := io.ReadAll(r)
	if err != nil {
		return Table{}, err
	}
	if len(raw) < 2 || raw[0] != 0xff || raw[1] != 0xfe {
		return Table{}, fmt.Errorf("xls: csv missing UTF-16LE BOM")
	}
	if (len(raw)-2)%2 != 0 {
		return Table{}, fmt.Errorf("xls: csv UTF-16 payload has odd byte length")
	}
	n := (len(raw) - 2) / 2
	u16 := make([]uint16, n)
	for i := range u16 {
		u16[i] = binary.LittleEndian.Uint16(raw[2+i*2 : 4+i*2])
	}
	s := string(utf16.Decode(u16))
	s = strings.ReplaceAll(s, "\r\n", "\n")
	idx := strings.IndexByte(s, '\n')
	if idx < 0 {
		return Table{}, fmt.Errorf("xls: csv missing newline after sep= line")
	}
	first := s[:idx]
	const sepPrefix = "sep="
	if !strings.HasPrefix(first, sepPrefix) {
		return Table{}, fmt.Errorf("xls: csv missing sep= line")
	}
	sepVal := first[len(sepPrefix):]
	sepRunes := []rune(sepVal)
	if len(sepRunes) != 1 || sepRunes[0] != delimR {
		return Table{}, fmt.Errorf("xls: sep= value %q does not match delimiter %q", sepVal, delimiter)
	}
	rest := s[idx+1:]
	cr := csv.NewReader(strings.NewReader(rest))
	cr.Comma = delimR
	cr.LazyQuotes = true
	records, err := cr.ReadAll()
	if err != nil {
		return Table{}, err
	}
	if len(records) == 0 {
		return Table{}, nil
	}
	if firstRowAsHeader {
		return Table{
			Columns: append([]string(nil), records[0]...),
			Rows:    append([][]string(nil), records[1:]...),
		}, nil
	}
	tab := Table{Rows: append([][]string(nil), records...)}
	return normalizeTable(tab), nil
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
