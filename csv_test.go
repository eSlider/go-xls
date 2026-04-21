package xls

import (
	"bytes"
	"io"
	"testing"
)

// mustRoundTripCSV writes tab with WriteCSV(..., writeDetectHead), then ReadCSV with readFirstRowAsHeader.
func mustRoundTripCSV(t *testing.T, tab Table, delimiter, enclosure, encoding string, writeDetectHead, readFirstRowAsHeader bool) Table {
	t.Helper()
	var buf bytes.Buffer
	if err := WriteCSV(&buf, tab, delimiter, enclosure, encoding, writeDetectHead); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}
	got, err := ReadCSV(bytes.NewReader(buf.Bytes()), delimiter, enclosure, readFirstRowAsHeader)
	if err != nil {
		t.Fatalf("ReadCSV: %v", err)
	}
	return got
}

func TestCSV_RoundTrip_WithHeader(t *testing.T) {
	want := Table{
		Columns: []string{"a", "b"},
		Rows: [][]string{
			{"1", "two,three"},
		},
	}
	got := mustRoundTripCSV(t, want, ",", `"`, "UTF-8", true, true)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestCSV_RoundTrip_Empty(t *testing.T) {
	got := mustRoundTripCSV(t, Table{}, ",", `"`, "UTF-8", true, true)
	if len(got.Columns) != 0 || len(got.Rows) != 0 {
		t.Fatalf("got %#v", got)
	}
}

func TestCSV_RoundTrip_NoHeaderRow(t *testing.T) {
	want := Table{
		Columns: []string{"0", "1"},
		Rows: [][]string{
			{"x", "y"},
			{"1", "2"},
		},
	}
	got := mustRoundTripCSV(t, want, ",", `"`, "UTF-8", false, false)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestCSV_RoundTrip_Latin1SupplementInCell(t *testing.T) {
	want := Table{
		Columns: []string{"word"},
		Rows:    [][]string{{"caf\u00e9"}},
	}
	got := mustRoundTripCSV(t, want, ",", `"`, "UTF-8", true, true)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestCSV_RoundTrip_SemicolonDelimiter(t *testing.T) {
	want := Table{
		Columns: []string{"a", "b"},
		Rows:    [][]string{{"1;2", "3"}},
	}
	got := mustRoundTripCSV(t, want, ";", `"`, "UTF-8", true, true)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestCSV_RoundTrip_SingleColumn(t *testing.T) {
	want := Table{
		Columns: []string{"x"},
		Rows:    [][]string{{"v"}},
	}
	got := mustRoundTripCSV(t, want, ",", `"`, "UTF-8", true, true)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestCSV_Write_UnsupportedEncoding(t *testing.T) {
	err := WriteCSV(io.Discard, Table{}, ",", `"`, "windows-1252", false)
	if err == nil {
		t.Fatal("expected error")
	}
}
