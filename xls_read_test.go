package xls

import (
	"bytes"
	"errors"
	"testing"
)

func TestParseXLS_RoundTripWithHeader(t *testing.T) {
	want := Table{
		Columns: []string{"name", "qty"},
		Rows: [][]string{
			{"apple", "3"},
			{"banana", "x"},
		},
	}
	b, err := GenXLS(want, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseXLS(b, true)
	if err != nil {
		t.Fatal(err)
	}
	if !tablesEqual(got, want) {
		t.Fatalf("parse mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestParseXLS_RoundTripNumericKeysNoHeader(t *testing.T) {
	want := Table{
		Columns: []string{"0", "1"},
		Rows: [][]string{
			{"a", "2"},
		},
	}
	b, err := GenXLS(want, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseXLS(b, false)
	if err != nil {
		t.Fatal(err)
	}
	if !tablesEqual(got, want) {
		t.Fatalf("parse mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestParseXLSToMaps(t *testing.T) {
	tab := Table{
		Columns: []string{"a", "b"},
		Rows:    [][]string{{"1", "2"}},
	}
	b, err := GenXLS(tab, true)
	if err != nil {
		t.Fatal(err)
	}
	maps, err := ParseXLSToMaps(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(maps) != 1 {
		t.Fatalf("len=%d", len(maps))
	}
	if maps[0]["a"] != "1" || maps[0]["b"] != "2" {
		t.Fatalf("%#v", maps[0])
	}
}

func TestParseXLS_OLERejected(t *testing.T) {
	b := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
	_, err := ParseXLS(b, true)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrOLEWorkbook) {
		t.Fatalf("err=%v", err)
	}
}

func TestParseXLS_Empty(t *testing.T) {
	b, err := GenXLS(Table{}, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseXLS(b, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Columns) != 0 || len(got.Rows) != 0 {
		t.Fatalf("%#v", got)
	}
}

func tablesEqual(a, b Table) bool {
	if len(a.Columns) != len(b.Columns) {
		return false
	}
	for i := range a.Columns {
		if a.Columns[i] != b.Columns[i] {
			return false
		}
	}
	if len(a.Rows) != len(b.Rows) {
		return false
	}
	for i := range a.Rows {
		if len(a.Rows[i]) != len(b.Rows[i]) {
			return false
		}
		for j := range a.Rows[i] {
			if a.Rows[i][j] != b.Rows[i][j] {
				return false
			}
		}
	}
	return true
}

func TestParseXLS_TrailingGarbageErrors(t *testing.T) {
	b, _ := GenXLS(Table{Columns: []string{"a"}, Rows: [][]string{{"b"}}}, true)
	b = append(b, []byte{0xFF, 0xFF}...)
	_, err := ParseXLS(b, true)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCellsToDenseGrid_DuplicateCell(t *testing.T) {
	_, err := cellsToDenseGrid([]sparseCell{
		{row: 0, col: 0, text: "a"},
		{row: 0, col: 0, text: "b"},
	})
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestIso88591RoundTrip(t *testing.T) {
	s := "caf\u00e9"
	b, err := toISO88591(s)
	if err != nil {
		t.Fatal(err)
	}
	if got := iso88591BytesToString(b); got != s {
		t.Fatalf("%q vs %q", got, s)
	}
}

func TestParseXLS_StreamWithoutEOF(t *testing.T) {
	// Minimal stream: BOF only (12 bytes) — no EOF, parser accepts if bytes exhausted after loop
	var buf bytes.Buffer
	writeU16LE(&buf, 0x809)
	writeU16LE(&buf, 0x8)
	writeU16LE(&buf, 0)
	writeU16LE(&buf, 0x10)
	writeU16LE(&buf, 0)
	writeU16LE(&buf, 0)
	_, err := ParseXLS(buf.Bytes(), true)
	if err != nil {
		t.Fatal(err)
	}
}
