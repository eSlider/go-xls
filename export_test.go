package xls

import (
	"bytes"
	"encoding/binary"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestWriteXLS_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteXLS(&buf, Table{}, true); err != nil {
		t.Fatal(err)
	}
	b := buf.Bytes()
	if len(b) < 14 {
		t.Fatalf("short output: %d bytes", len(b))
	}
	wantBOFPrefix := []byte{0x09, 0x08, 0x08, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00}
	if !bytes.HasPrefix(b, wantBOFPrefix) {
		t.Fatalf("unexpected BOF prefix: %#v", b[:12])
	}
	if !bytes.HasSuffix(b, []byte{0x0a, 0x00, 0x00, 0x00}) {
		t.Fatalf("unexpected EOF suffix: %#v", b[len(b)-4:])
	}
}

func TestWriteXLS_HeaderAndRows(t *testing.T) {
	tab := Table{
		Columns: []string{"name", "qty"},
		Rows: [][]string{
			{"apple", "3"},
			{"banana", "x"},
		},
	}
	var buf bytes.Buffer
	if err := WriteXLS(&buf, tab, true); err != nil {
		t.Fatal(err)
	}
	b := buf.Bytes()
	if !bytes.HasPrefix(b, []byte{0x09, 0x08, 0x08, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		t.Fatalf("bad BOF")
	}
	if !containsSubseq(b, buildNumberRecord(1, 1, 3)) {
		t.Fatalf("missing numeric cell for qty=3: sample %#v", b[:min(80, len(b))])
	}
}

func TestWriteXLS_NumericKeysNoHeader(t *testing.T) {
	tab := Table{
		Columns: []string{"0", "1"},
		Rows: [][]string{
			{"a", "2"},
		},
	}
	var buf bytes.Buffer
	if err := WriteXLS(&buf, tab, true); err != nil {
		t.Fatal(err)
	}
	b := buf.Bytes()
	if !containsSubseq(b, buildNumberRecord(0, 1, 2)) {
		t.Fatalf("expected numeric cell at row0 col1")
	}
}

func buildNumberRecord(row, col uint16, v float64) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint16(XLSIntType))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(14))
	_ = binary.Write(&buf, binary.LittleEndian, row)
	_ = binary.Write(&buf, binary.LittleEndian, col)
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0))
	_ = binary.Write(&buf, binary.LittleEndian, v)
	return buf.Bytes()
}

func containsSubseq(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if bytes.Equal(haystack[i:i+len(needle)], needle) {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestWriteCSV_BOMSepAndRoundTrip(t *testing.T) {
	tab := Table{
		Columns: []string{"a", "b"},
		Rows: [][]string{
			{"1", "two,three"},
		},
	}
	var buf bytes.Buffer
	if err := WriteCSV(&buf, tab, ",", `"`, "UTF-8", true); err != nil {
		t.Fatal(err)
	}
	b := buf.Bytes()
	if len(b) < 2 || b[0] != 0xff || b[1] != 0xfe {
		t.Fatalf("missing utf-16le bom: %#v", b[:4])
	}
	u16 := make([]uint16, (len(b)-2)/2)
	for i := range u16 {
		u16[i] = uint16(b[2+i*2]) | uint16(b[3+i*2])<<8
	}
	s := string(utf16.Decode(u16))
	if !strings.HasPrefix(s, "sep=,\n") {
		t.Fatalf("missing sep line: %q", s[:min(20, len(s))])
	}
	if !strings.Contains(s, `"two,three"`) {
		t.Fatalf("expected quoted field, got: %q", s)
	}
}

func TestWriteCSV_UnsupportedEncoding(t *testing.T) {
	err := WriteCSV(io.Discard, Table{}, ",", `"`, "windows-1252", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeTable(t *testing.T) {
	tab := normalizeTable(Table{Rows: [][]string{{"a", "b", "c"}, {"d"}}})
	if got, want := strings.Join(tab.Columns, ","), "0,1,2"; got != want {
		t.Fatalf("columns=%q want %q", got, want)
	}
}

func TestWriteAttachment(t *testing.T) {
	rec := httptest.NewRecorder()
	body := strings.NewReader("abc")
	if err := WriteAttachment(rec, "export.xls", ContentTypeXLS, body, 3, true); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	defer res.Body.Close()
	if res.Header.Get("Content-Type") != ContentTypeXLS {
		t.Fatalf("content-type=%q", res.Header.Get("Content-Type"))
	}
	if !strings.Contains(res.Header.Get("Content-Disposition"), `filename="export.xls"`) {
		t.Fatalf("disposition=%q", res.Header.Get("Content-Disposition"))
	}
	if res.Header.Get("Cache-Control") != "private" {
		t.Fatalf("cache-control=%q", res.Header.Get("Cache-Control"))
	}
	out, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "abc" {
		t.Fatalf("body=%q", out)
	}
}
