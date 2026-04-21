package xls

import (
	"strings"
	"testing"
)

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

func TestNormalizeTable(t *testing.T) {
	tab := normalizeTable(Table{Rows: [][]string{{"a", "b", "c"}, {"d"}}})
	if got, want := strings.Join(tab.Columns, ","), "0,1,2"; got != want {
		t.Fatalf("columns=%q want %q", got, want)
	}
}
