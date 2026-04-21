package xls

import (
	"fmt"
	"strings"
)

// MarkdownAlign is column alignment in a GitHub-style pipe table divider row.
type MarkdownAlign int

const (
	AlignLeft MarkdownAlign = iota
	AlignCenter
	AlignRight
)

// MarkdownMarshalOpts controls markdown table output from [MarshalMarkdownTableWith].
type MarkdownMarshalOpts struct {
	// Align is per-column alignment (shorter slice defaults remaining columns to left).
	Align []MarkdownAlign
}

// MarshalMarkdownTable renders tab as a GitHub-flavored markdown pipe table.
// Header row uses [Table.Columns] (after [normalizeTable] if columns were empty).
func MarshalMarkdownTable(tab Table) (string, error) {
	return MarshalMarkdownTableWith(tab, MarkdownMarshalOpts{})
}

// MarshalMarkdownTableWith renders tab as a markdown pipe table with optional column alignment.
func MarshalMarkdownTableWith(tab Table, opts MarkdownMarshalOpts) (string, error) {
	tab = normalizeTable(tab)
	if len(tab.Columns) == 0 {
		return "", fmt.Errorf("xls: markdown table needs at least one column")
	}

	var b strings.Builder
	writeMarkdownPipeRow(&b, tab.Columns)
	b.WriteByte('\n')
	writeMarkdownDivider(&b, len(tab.Columns), opts.Align)
	b.WriteByte('\n')
	for _, row := range tab.Rows {
		cells := make([]string, len(tab.Columns))
		for i := range tab.Columns {
			if i < len(row) {
				cells[i] = row[i]
			}
		}
		writeMarkdownPipeRow(&b, cells)
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n"), nil
}

func writeMarkdownPipeRow(w *strings.Builder, cells []string) {
	w.WriteByte('|')
	for _, c := range cells {
		w.WriteByte(' ')
		w.WriteString(escapeMarkdownCell(c))
		w.WriteString(" |")
	}
}

func writeMarkdownDivider(w *strings.Builder, ncols int, align []MarkdownAlign) {
	w.WriteByte('|')
	for i := 0; i < ncols; i++ {
		a := AlignLeft
		if i < len(align) {
			a = align[i]
		}
		w.WriteByte(' ')
		w.WriteString(markdownDividerField(a))
		w.WriteString(" |")
	}
}

func markdownDividerField(a MarkdownAlign) string {
	switch a {
	case AlignCenter:
		return ":---:"
	case AlignRight:
		return "---:"
	default:
		return ":---"
	}
}

func escapeMarkdownCell(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func unescapeMarkdownCell(s string) string {
	rs := []rune(s)
	var b strings.Builder
	for i := 0; i < len(rs); i++ {
		if rs[i] == '\\' && i+1 < len(rs) {
			switch rs[i+1] {
			case '|':
				b.WriteRune('|')
				i++
				continue
			case '\\':
				b.WriteRune('\\')
				i++
				continue
			}
		}
		b.WriteRune(rs[i])
	}
	return b.String()
}

// UnmarshalMarkdownTable parses the first GitHub-style pipe table in s (header + divider + body rows).
// Leading / trailing blank lines are ignored. Prose lines before the first header+divider pair are skipped.
func UnmarshalMarkdownTable(s string) (Table, error) {
	tab, _, err := UnmarshalMarkdownTableDetailed(s)
	return tab, err
}

// UnmarshalMarkdownTableDetailed is like [UnmarshalMarkdownTable] but also returns column alignment
// decoded from the divider row (for round-tripping with [MarshalMarkdownTableWith]).
func UnmarshalMarkdownTableDetailed(s string) (Table, []MarkdownAlign, error) {
	lines := splitMarkdownLines(s)
	start := -1
	for i, ln := range lines {
		if isProbablyTableRow(ln) && i+1 < len(lines) && isMarkdownDividerRow(lines[i+1]) {
			start = i
			break
		}
	}
	if start < 0 {
		return Table{}, nil, fmt.Errorf("xls: no markdown pipe table found")
	}

	headerCells, err := splitMarkdownRow(lines[start])
	if err != nil {
		return Table{}, nil, err
	}
	if len(headerCells) == 0 {
		return Table{}, nil, fmt.Errorf("xls: empty markdown header row")
	}

	align, err := parseMarkdownDivider(lines[start+1], len(headerCells))
	if err != nil {
		return Table{}, nil, err
	}

	var rows [][]string
	for _, ln := range lines[start+2:] {
		if strings.TrimSpace(ln) == "" {
			break
		}
		if isMarkdownDividerRow(ln) {
			break
		}
		if !isProbablyTableRow(ln) {
			break
		}
		cells, err := splitMarkdownRow(ln)
		if err != nil {
			return Table{}, nil, err
		}
		if len(cells) != len(headerCells) {
			return Table{}, nil, fmt.Errorf("xls: markdown row width %d != header %d", len(cells), len(headerCells))
		}
		rows = append(rows, cells)
	}

	return Table{Columns: headerCells, Rows: rows}, align, nil
}

func splitMarkdownLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

func isProbablyTableRow(line string) bool {
	line = strings.TrimSpace(line)
	return strings.Contains(line, "|")
}

func isMarkdownDividerRow(line string) bool {
	cells, err := splitMarkdownRow(line)
	if err != nil || len(cells) == 0 {
		return false
	}
	for _, c := range cells {
		if !isDividerField(c) {
			return false
		}
	}
	return true
}

func isDividerField(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	lhs := strings.HasPrefix(s, ":")
	rhs := strings.HasSuffix(s, ":")
	core := s
	if lhs {
		core = core[1:]
	}
	if rhs && len(core) > 0 {
		core = core[:len(core)-1]
	}
	for _, r := range core {
		if r != '-' && r != ' ' && r != '\t' {
			return false
		}
	}
	return strings.Contains(core, "-")
}

func parseMarkdownDivider(line string, wantCols int) ([]MarkdownAlign, error) {
	cells, err := splitMarkdownRow(line)
	if err != nil {
		return nil, err
	}
	if len(cells) != wantCols {
		return nil, fmt.Errorf("xls: divider columns %d != header %d", len(cells), wantCols)
	}
	out := make([]MarkdownAlign, len(cells))
	for i, c := range cells {
		out[i] = alignFromDividerField(c)
	}
	return out, nil
}

func alignFromDividerField(s string) MarkdownAlign {
	s = strings.TrimSpace(s)
	lhs := strings.HasPrefix(s, ":")
	rhs := strings.HasSuffix(s, ":")
	if lhs && rhs {
		return AlignCenter
	}
	if rhs {
		return AlignRight
	}
	return AlignLeft
}

func splitMarkdownRow(line string) ([]string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("xls: empty markdown row")
	}
	rs := []rune(line)
	if len(rs) > 0 && rs[0] == '|' {
		rs = rs[1:]
	}
	if len(rs) > 0 && rs[len(rs)-1] == '|' {
		rs = rs[:len(rs)-1]
	}
	var parts []string
	var cur strings.Builder
	for i := 0; i < len(rs); i++ {
		if rs[i] == '|' && !escapedBeforeRune(rs, i) {
			parts = append(parts, strings.TrimSpace(unescapeMarkdownCell(cur.String())))
			cur.Reset()
			continue
		}
		cur.WriteRune(rs[i])
	}
	parts = append(parts, strings.TrimSpace(unescapeMarkdownCell(cur.String())))
	return parts, nil
}

func escapedBeforeRune(rs []rune, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && rs[j] == '\\'; j-- {
		n++
	}
	return n%2 == 1
}
