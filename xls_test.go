package xls

import (
	"bytes"
	"errors"
	"testing"

	"github.com/eslider/go-xls/v2/biff"
	"github.com/eslider/go-xls/v2/ole"
)

// mustRoundTripXLS writes tab with WriteXLS(..., writeDetectHead), then ReadXLS with readFirstRowAsHeader.
// It asserts the full encode → decode loop for linear BIFF .xls.
func mustRoundTripXLS(t *testing.T, tab Table, writeDetectHead, readFirstRowAsHeader bool) Table {
	t.Helper()
	var buf bytes.Buffer
	if err := WriteXLS(&buf, tab, writeDetectHead); err != nil {
		t.Fatalf("WriteXLS: %v", err)
	}
	got, err := ReadXLS(bytes.NewReader(buf.Bytes()), readFirstRowAsHeader)
	if err != nil {
		t.Fatalf("ReadXLS: %v", err)
	}
	return got
}

func TestXLS_RoundTrip_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteXLS(&buf, Table{}, true); err != nil {
		t.Fatal(err)
	}
	if buf.Len() < 14 {
		t.Fatalf("output too short: %d bytes", buf.Len())
	}
	got, err := ReadXLS(bytes.NewReader(buf.Bytes()), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Columns) != 0 || len(got.Rows) != 0 {
		t.Fatalf("got %#v", got)
	}
}

func TestXLS_RoundTrip_WithHeader(t *testing.T) {
	want := Table{
		Columns: []string{"name", "qty"},
		Rows: [][]string{
			{"apple", "3"},
			{"banana", "x"},
		},
	}
	got := mustRoundTripXLS(t, want, true, true)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestXLS_RoundTrip_NumericKeysNoHeader(t *testing.T) {
	want := Table{
		Columns: []string{"0", "1"},
		Rows: [][]string{
			{"a", "2"},
		},
	}
	got := mustRoundTripXLS(t, want, true, false)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestXLS_RoundTrip_ToMaps(t *testing.T) {
	tab := Table{
		Columns: []string{"a", "b"},
		Rows:    [][]string{{"1", "2"}},
	}
	var buf bytes.Buffer
	if err := WriteXLS(&buf, tab, true); err != nil {
		t.Fatalf("WriteXLS: %v", err)
	}
	maps, err := ReadXLSToMaps(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ReadXLSToMaps: %v", err)
	}
	if len(maps) != 1 {
		t.Fatalf("len=%d", len(maps))
	}
	if maps[0]["a"] != "1" || maps[0]["b"] != "2" {
		t.Fatalf("%#v", maps[0])
	}
}

func TestXLS_RoundTrip_Latin1SupplementInCell(t *testing.T) {
	want := Table{
		Columns: []string{"word"},
		Rows:    [][]string{{"caf\u00e9"}},
	}
	got := mustRoundTripXLS(t, want, true, true)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch (ISO-8859-1 path)\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestXLS_RoundTrip_FloatCell(t *testing.T) {
	want := Table{
		Columns: []string{"v"},
		Rows:    [][]string{{"1.25"}, {"-0"}},
	}
	got := mustRoundTripXLS(t, want, true, true)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestReadXLS_OLERejected(t *testing.T) {
	b := append(append([]byte{}, ole.HeaderPrefix[:]...), 0xA1, 0xB1, 0x1A, 0xE1)
	_, err := ReadXLS(bytes.NewReader(b), true)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrOLEWorkbook) {
		t.Fatalf("err=%v", err)
	}
}

func TestXLS_WriteThenRead_TrailingBytesRejected(t *testing.T) {
	var wbuf bytes.Buffer
	if err := WriteXLS(&wbuf, Table{Columns: []string{"a"}, Rows: [][]string{{"b"}}}, true); err != nil {
		t.Fatal(err)
	}
	b := append(wbuf.Bytes(), []byte{0xFF, 0xFF}...)
	_, err := ReadXLS(bytes.NewReader(b), true)
	if err == nil {
		t.Fatal("expected error after valid stream")
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

func TestReadXLS_BOFOnlyStream(t *testing.T) {
	var buf bytes.Buffer
	writeU16LE(&buf, biff.RecordBOF)
	writeU16LE(&buf, biff.BOFPayloadLen)
	writeU16LE(&buf, biff.BOFVersionMinor)
	writeU16LE(&buf, biff.BOFVersionMajor)
	writeU16LE(&buf, biff.BOFReserved0)
	writeU16LE(&buf, biff.BOFReserved1)
	_, err := ReadXLS(bytes.NewReader(buf.Bytes()), true)
	if err != nil {
		t.Fatal(err)
	}
}
