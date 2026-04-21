package xls

import (
	"fmt"
	"io"
	"strings"
)

// MarkdownAlign is column alignment in a GitHub-style pipe table divider row.
type MarkdownAlign int

const (
	AlignLeft MarkdownAlign = iota
	AlignCenter
	AlignRight
)

// MarkdownMarshalOpts controls markdown table output from [WriteMarkdownTableWith].
type MarkdownMarshalOpts struct {
	// Align is per-column alignment (shorter slice defaults remaining columns to left).
	Align []MarkdownAlign
}

// WriteMarkdownTable writes a GitHub-flavored markdown pipe table to w.
// The header row uses [Table.Columns] (after [normalizeTable] if columns were empty).
func WriteMarkdownTable(w io.Writer, tab Table) error {
	return WriteMarkdownTableWith(w, tab, MarkdownMarshalOpts{})
}

// WriteMarkdownTableWith writes a markdown pipe table to w with optional column alignment.
func WriteMarkdownTableWith(w io.Writer, tab Table, opts MarkdownMarshalOpts) error {
	tab = normalizeTable(tab)
	if len(tab.Columns) == 0 {
		return fmt.Errorf("xls: markdown table needs at least one column")
	}

	if err := writeMarkdownPipeRow(w, tab.Columns); err != nil {
		return err
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}
	if err := writeMarkdownDivider(w, len(tab.Columns), opts.Align); err != nil {
		return err
	}
	for _, row := range tab.Rows {
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
		cells := make([]string, len(tab.Columns))
		for j := range tab.Columns {
			if j < len(row) {
				cells[j] = row[j]
			}
		}
		if err := writeMarkdownPipeRow(w, cells); err != nil {
			return err
		}
	}
	return nil
}

func writeMarkdownPipeRow(w io.Writer, cells []string) error {
	if _, err := w.Write([]byte{'|'}); err != nil {
		return err
	}
	for _, c := range cells {
		if _, err := fmt.Fprintf(w, " %s |", escapeMarkdownCell(c)); err != nil {
			return err
		}
	}
	return nil
}

func writeMarkdownDivider(w io.Writer, ncols int, align []MarkdownAlign) error {
	if _, err := w.Write([]byte{'|'}); err != nil {
		return err
	}
	for i := 0; i < ncols; i++ {
		a := AlignLeft
		if i < len(align) {
			a = align[i]
		}
		if _, err := fmt.Fprintf(w, " %s |", markdownDividerField(a)); err != nil {
			return err
		}
	}
	return nil
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

// ReadMarkdownTable parses the first GitHub-style pipe table from r (header + divider + body rows).
// Prose before the first header+divider pair is skipped.
func ReadMarkdownTable(r io.Reader) (Table, error) {
	tab, _, err := ReadMarkdownTableDetailed(r)
	return tab, err
}

// ReadMarkdownTableDetailed is like [ReadMarkdownTable] but also returns column alignment
// from the divider row (for round-tripping with [WriteMarkdownTableWith]).
func ReadMarkdownTableDetailed(r io.Reader) (Table, []MarkdownAlign, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Table{}, nil, err
	}
	return parseMarkdownTableFromString(string(data))
}

func parseMarkdownTableFromString(s string) (Table, []MarkdownAlign, error) {
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
