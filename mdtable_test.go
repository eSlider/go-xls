package xls

import (
	"bytes"
	"strings"
	"testing"
)

// mustRoundTripMarkdown writes want with [WriteMarkdownTable], then [ReadMarkdownTable].
func mustRoundTripMarkdown(t *testing.T, want Table) Table {
	t.Helper()
	var w bytes.Buffer
	if err := WriteMarkdownTable(&w, want); err != nil {
		t.Fatalf("WriteMarkdownTable: %v", err)
	}
	got, err := ReadMarkdownTable(bytes.NewReader(w.Bytes()))
	if err != nil {
		t.Fatalf("ReadMarkdownTable: %v", err)
	}
	return got
}

// mustRoundTripMarkdownWith writes with [WriteMarkdownTableWith], then [ReadMarkdownTableDetailed].
func mustRoundTripMarkdownWith(t *testing.T, tab Table, opts MarkdownMarshalOpts) (Table, []MarkdownAlign) {
	t.Helper()
	var w bytes.Buffer
	if err := WriteMarkdownTableWith(&w, tab, opts); err != nil {
		t.Fatalf("WriteMarkdownTableWith: %v", err)
	}
	gotTab, gotAlign, err := ReadMarkdownTableDetailed(bytes.NewReader(w.Bytes()))
	if err != nil {
		t.Fatalf("ReadMarkdownTableDetailed: %v", err)
	}
	return gotTab, gotAlign
}

func TestMarkdown_RoundTrip_Basic(t *testing.T) {
	want := Table{
		Columns: []string{"name", "qty"},
		Rows: [][]string{
			{"apple", "3"},
			{"banana", "x"},
		},
	}
	got := mustRoundTripMarkdown(t, want)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestMarkdown_RoundTrip_PipeInCell(t *testing.T) {
	want := Table{
		Columns: []string{"a"},
		Rows:    [][]string{{"p|q"}},
	}
	got := mustRoundTripMarkdown(t, want)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestMarkdown_RoundTrip_BackslashInCell(t *testing.T) {
	want := Table{
		Columns: []string{"x"},
		Rows:    [][]string{{`C:\temp`}},
	}
	got := mustRoundTripMarkdown(t, want)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestMarkdown_RoundTrip_AfterProsePrefix(t *testing.T) {
	want := Table{Columns: []string{"a", "b"}, Rows: [][]string{{"1", "2"}}}
	var w bytes.Buffer
	if err := WriteMarkdownTable(&w, want); err != nil {
		t.Fatalf("WriteMarkdownTable: %v", err)
	}
	doc := "Some intro\n\n" + w.String()
	got, err := ReadMarkdownTable(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("ReadMarkdownTable: %v", err)
	}
	if !tablesEqual(got, want) {
		t.Fatalf("write→read (with prose prefix) mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestMarkdown_RoundTrip_Alignment(t *testing.T) {
	tab := Table{
		Columns: []string{"L", "C", "R"},
		Rows:    [][]string{{"a", "b", "c"}},
	}
	opts := MarkdownMarshalOpts{
		Align: []MarkdownAlign{AlignLeft, AlignCenter, AlignRight},
	}
	gotTab, gotAlign := mustRoundTripMarkdownWith(t, tab, opts)
	if !tablesEqual(gotTab, tab) {
		t.Fatalf("write→read table mismatch %#v", gotTab)
	}
	if len(gotAlign) != 3 || gotAlign[0] != AlignLeft || gotAlign[1] != AlignCenter || gotAlign[2] != AlignRight {
		t.Fatalf("align %#v", gotAlign)
	}
	var w2 bytes.Buffer
	if err := WriteMarkdownTableWith(&w2, gotTab, MarkdownMarshalOpts{Align: gotAlign}); err != nil {
		t.Fatalf("remarshal: %v", err)
	}
	var w1 bytes.Buffer
	if err := WriteMarkdownTableWith(&w1, tab, opts); err != nil {
		t.Fatal(err)
	}
	if w2.String() != w1.String() {
		t.Fatalf("second write→read loop bytes differ:\n%s\nvs\n%s", w1.String(), w2.String())
	}
}

func TestMarkdown_RoundTrip_HeaderOnly(t *testing.T) {
	want := Table{Columns: []string{"x"}, Rows: nil}
	got := mustRoundTripMarkdown(t, want)
	if !tablesEqual(got, want) {
		t.Fatalf("write→read mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestMarkdown_Read_NoTable(t *testing.T) {
	_, err := ReadMarkdownTable(strings.NewReader("hello\nworld\n"))
	if err == nil {
		t.Fatal("expected error")
	}
}
