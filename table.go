package xls

import "strconv"

// Table is column-ordered tabular data for export.
// Columns defines order and names; each Rows[i] may be shorter than len(Columns) (missing cells are empty).
// If Columns is nil or empty, synthetic names "0","1",… are derived from the widest row (Mapbender-style numeric keys).
type Table struct {
	Columns []string
	Rows    [][]string
}

func normalizeTable(tab Table) Table {
	if len(tab.Columns) > 0 {
		return tab
	}
	max := 0
	for _, r := range tab.Rows {
		if len(r) > max {
			max = len(r)
		}
	}
	if max == 0 {
		return tab
	}
	cols := make([]string, max)
	for i := range cols {
		cols[i] = strconv.Itoa(i)
	}
	return Table{Columns: cols, Rows: tab.Rows}
}
