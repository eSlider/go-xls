package xls

import (
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
	md, err := MarshalMarkdownTable(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalMarkdownTable(md)
	if err != nil {
		t.Fatal(err)
	}
	if !tablesEqual(got, want) {
		t.Fatalf("got %#v want %#v\n--- md ---\n%s", got, want, md)
	}
}

func TestMarkdownPipeInCell(t *testing.T) {
	want := Table{
		Columns: []string{"a"},
		Rows:    [][]string{{"p|q"}},
	}
	md, err := MarshalMarkdownTable(want)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, `\|`) {
		t.Fatalf("expected escaped pipe in md: %q", md)
	}
	got, err := UnmarshalMarkdownTable(md)
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
	md, err := MarshalMarkdownTable(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalMarkdownTable(md)
	if err != nil {
		t.Fatal(err)
	}
	if !tablesEqual(got, want) {
		t.Fatalf("got %#v md=%q", got, md)
	}
}

func TestMarkdownProseBeforeTable(t *testing.T) {
	s := "Some intro\n\n| a | b |\n| --- | --- |\n| 1 | 2 |\n"
	got, err := UnmarshalMarkdownTable(s)
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
	want := MarkdownMarshalOpts{
		Align: []MarkdownAlign{AlignLeft, AlignCenter, AlignRight},
	}
	md, err := MarshalMarkdownTableWith(tab, want)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, ":---:") || !strings.Contains(md, "---:") {
		t.Fatalf("md=%q", md)
	}
	gotTab, gotAlign, err := UnmarshalMarkdownTableDetailed(md)
	if err != nil {
		t.Fatal(err)
	}
	if !tablesEqual(gotTab, tab) {
		t.Fatalf("table %#v", gotTab)
	}
	if len(gotAlign) != 3 || gotAlign[0] != AlignLeft || gotAlign[1] != AlignCenter || gotAlign[2] != AlignRight {
		t.Fatalf("align %#v", gotAlign)
	}
	md2, err := MarshalMarkdownTableWith(gotTab, MarkdownMarshalOpts{Align: gotAlign})
	if err != nil {
		t.Fatal(err)
	}
	if md2 != md {
		t.Fatalf("remarshal differs:\n%s\nvs\n%s", md, md2)
	}
}

func TestUnmarshalMarkdownTable_NoTable(t *testing.T) {
	_, err := UnmarshalMarkdownTable("hello\nworld\n")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalMarkdownTable_HeaderOnly(t *testing.T) {
	md := "| x |\n| --- |\n"
	tab, err := UnmarshalMarkdownTable(md)
	if err != nil {
		t.Fatal(err)
	}
	if len(tab.Columns) != 1 || tab.Columns[0] != "x" || len(tab.Rows) != 0 {
		t.Fatalf("%#v", tab)
	}
}
