package xls

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarkdownRoundTrip(t *testing.T) {
	want := Table{
		Columns: []string{"name", "qty"},
		Rows: [][]string{
			{"apple", "3"},
			{"banana", "x"},
		},
	}
	var w bytes.Buffer
	if err := WriteMarkdownTable(&w, want); err != nil {
		t.Fatal(err)
	}
	got, err := ReadMarkdownTable(bytes.NewReader(w.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if !tablesEqual(got, want) {
		t.Fatalf("got %#v want %#v\n--- md ---\n%s", got, want, w.String())
	}
}

func TestMarkdownPipeInCell(t *testing.T) {
	want := Table{
		Columns: []string{"a"},
		Rows:    [][]string{{"p|q"}},
	}
	var w bytes.Buffer
	if err := WriteMarkdownTable(&w, want); err != nil {
		t.Fatal(err)
	}
	md := w.String()
	if !strings.Contains(md, `\|`) {
		t.Fatalf("expected escaped pipe in md: %q", md)
	}
	got, err := ReadMarkdownTable(strings.NewReader(md))
	if err != nil {
		t.Fatal(err)
	}
	if !tablesEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
}

func TestMarkdownBackslashInCell(t *testing.T) {
	want := Table{
		Columns: []string{"x"},
		Rows:    [][]string{{`C:\temp`}},
	}
	var w bytes.Buffer
	if err := WriteMarkdownTable(&w, want); err != nil {
		t.Fatal(err)
	}
	got, err := ReadMarkdownTable(bytes.NewReader(w.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if !tablesEqual(got, want) {
		t.Fatalf("got %#v md=%q", got, w.String())
	}
}

func TestMarkdownProseBeforeTable(t *testing.T) {
	s := "Some intro\n\n| a | b |\n| --- | --- |\n| 1 | 2 |\n"
	got, err := ReadMarkdownTable(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	want := Table{Columns: []string{"a", "b"}, Rows: [][]string{{"1", "2"}}}
	if !tablesEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
}

func TestMarkdownAlignmentRoundTrip(t *testing.T) {
	tab := Table{
		Columns: []string{"L", "C", "R"},
		Rows:    [][]string{{"a", "b", "c"}},
	}
	opts := MarkdownMarshalOpts{
		Align: []MarkdownAlign{AlignLeft, AlignCenter, AlignRight},
	}
	var w1 bytes.Buffer
	if err := WriteMarkdownTableWith(&w1, tab, opts); err != nil {
		t.Fatal(err)
	}
	md := w1.String()
	if !strings.Contains(md, ":---:") || !strings.Contains(md, "---:") {
		t.Fatalf("md=%q", md)
	}
	gotTab, gotAlign, err := ReadMarkdownTableDetailed(strings.NewReader(md))
	if err != nil {
		t.Fatal(err)
	}
	if !tablesEqual(gotTab, tab) {
		t.Fatalf("table %#v", gotTab)
	}
	if len(gotAlign) != 3 || gotAlign[0] != AlignLeft || gotAlign[1] != AlignCenter || gotAlign[2] != AlignRight {
		t.Fatalf("align %#v", gotAlign)
	}
	var w2 bytes.Buffer
	if err := WriteMarkdownTableWith(&w2, gotTab, MarkdownMarshalOpts{Align: gotAlign}); err != nil {
		t.Fatal(err)
	}
	if w2.String() != md {
		t.Fatalf("remarshal differs:\n%s\nvs\n%s", md, w2.String())
	}
}

func TestReadMarkdownTable_NoTable(t *testing.T) {
	_, err := ReadMarkdownTable(strings.NewReader("hello\nworld\n"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadMarkdownTable_HeaderOnly(t *testing.T) {
	md := "| x |\n| --- |\n"
	tab, err := ReadMarkdownTable(strings.NewReader(md))
	if err != nil {
		t.Fatal(err)
	}
	if len(tab.Columns) != 1 || tab.Columns[0] != "x" || len(tab.Rows) != 0 {
		t.Fatalf("%#v", tab)
	}
}
