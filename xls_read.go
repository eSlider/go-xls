package xls

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
)

// BIFF record types for the linear stream read by [ReadXLS].
const (
	biffBOF = 0x809 // LE bytes 09 08
	biffEOF = 0x0A
)

const (
	maxBIFFRecordPayload = 1 << 20 // 1 MiB per record
	maxXLSCells          = 1_000_000
)

var (
	// ErrOLEWorkbook means the input is an OLE compound document (.xls as produced by Excel 97).
	// [ReadXLS] only supports the linear BIFF stream from [WriteXLS]; use another tool or convert to .xlsx.
	ErrOLEWorkbook = errors.New("xls: OLE compound workbook not supported (linear BIFF stream only)")
	// ErrTruncatedXLS means the stream ended inside a BIFF record.
	ErrTruncatedXLS = errors.New("xls: truncated record")
)

// sparseCell holds one decoded cell from the BIFF stream.
type sparseCell struct {
	row, col uint16
	text     string
}

// ReadXLS reads a legacy .xls linear BIFF stream from r. It is not a full OLE workbook parser;
// inputs starting with the OLE signature return [ErrOLEWorkbook].
//
// When firstRowAsHeader is true, row 0 becomes [Table.Columns] and following rows [Table.Rows].
// When false, columns are named "0","1",… (widest row) and every grid row is in [Table.Rows].
func ReadXLS(r io.Reader, firstRowAsHeader bool) (Table, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return Table{}, err
	}
	if len(b) >= 4 && b[0] == 0xD0 && b[1] == 0xCF && b[2] == 0x11 && b[3] == 0xE0 {
		return Table{}, ErrOLEWorkbook
	}

	cells, err := decodeXLSStream(b)
	if err != nil {
		return Table{}, err
	}
	grid, err := cellsToDenseGrid(cells)
	if err != nil {
		return Table{}, err
	}
	if len(grid) == 0 {
		return Table{}, nil
	}

	if firstRowAsHeader {
		return Table{
			Columns: append([]string(nil), grid[0]...),
			Rows:    cloneStringMatrix(grid[1:]),
		}, nil
	}
	tab := Table{Rows: cloneStringMatrix(grid)}
	return normalizeTable(tab), nil
}

// ReadXLSToMaps reads r like [ReadXLS] with a header row and returns one map per data row.
// Duplicate header names: later columns overwrite earlier keys in each map.
func ReadXLSToMaps(r io.Reader) ([]map[string]string, error) {
	t, err := ReadXLS(r, true)
	if err != nil {
		return nil, err
	}
	if len(t.Columns) == 0 {
		return nil, nil
	}
	out := make([]map[string]string, len(t.Rows))
	for i, row := range t.Rows {
		m := make(map[string]string, len(t.Columns))
		for j, colName := range t.Columns {
			if j < len(row) {
				m[colName] = row[j]
			} else {
				m[colName] = ""
			}
		}
		out[i] = m
	}
	return out, nil
}

func decodeXLSStream(b []byte) ([]sparseCell, error) {
	var cells []sparseCell
	off := 0
	for off+4 <= len(b) {
		typ := binary.LittleEndian.Uint16(b[off : off+2])
		reclen := int(binary.LittleEndian.Uint16(b[off+2 : off+4]))
		off += 4
		if reclen < 0 || reclen > maxBIFFRecordPayload {
			return nil, fmt.Errorf("xls: invalid record length %d for opcode %#x", reclen, typ)
		}
		if off+reclen > len(b) {
			return nil, ErrTruncatedXLS
		}
		payload := b[off : off+reclen]
		off += reclen

		switch typ {
		case biffBOF:
			// ignore BOF payload
		case biffEOF:
			if reclen != 0 {
				return nil, fmt.Errorf("xls: unexpected EOF payload length %d", reclen)
			}
			if off != len(b) {
				return nil, fmt.Errorf("xls: %d trailing bytes after EOF", len(b)-off)
			}
			return cells, nil
		case XLSStringType:
			c, err := parseStringRecord(payload)
			if err != nil {
				return nil, err
			}
			cells = append(cells, c)
			if len(cells) > maxXLSCells {
				return nil, fmt.Errorf("xls: too many cells (> %d)", maxXLSCells)
			}
		case XLSIntType:
			c, err := parseNumberRecord(payload)
			if err != nil {
				return nil, err
			}
			cells = append(cells, c)
			if len(cells) > maxXLSCells {
				return nil, fmt.Errorf("xls: too many cells (> %d)", maxXLSCells)
			}
		default:
			// skip unknown records
		}
	}
	if off < len(b) {
		if len(b)-off < 4 {
			return nil, fmt.Errorf("xls: incomplete record header (%d bytes left)", len(b)-off)
		}
		return nil, fmt.Errorf("xls: %d trailing bytes after last complete record", len(b)-off)
	}
	return cells, nil
}

func parseStringRecord(payload []byte) (sparseCell, error) {
	if len(payload) < 8 {
		return sparseCell{}, fmt.Errorf("xls: string record payload too short (%d)", len(payload))
	}
	row := binary.LittleEndian.Uint16(payload[0:2])
	col := binary.LittleEndian.Uint16(payload[2:4])
	strLen := int(binary.LittleEndian.Uint16(payload[6:8]))
	if 8+strLen != len(payload) {
		return sparseCell{}, fmt.Errorf("xls: string record length mismatch (declared %d, payload %d)", strLen, len(payload))
	}
	raw := payload[8:]
	s := iso88591BytesToString(raw)
	return sparseCell{row: row, col: col, text: s}, nil
}

func parseNumberRecord(payload []byte) (sparseCell, error) {
	if len(payload) != 14 {
		return sparseCell{}, fmt.Errorf("xls: number record expected 14-byte payload, got %d", len(payload))
	}
	row := binary.LittleEndian.Uint16(payload[0:2])
	col := binary.LittleEndian.Uint16(payload[2:4])
	v := math.Float64frombits(binary.LittleEndian.Uint64(payload[6:14]))
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return sparseCell{}, fmt.Errorf("xls: invalid numeric cell")
	}
	return sparseCell{row: row, col: col, text: formatXLSFloat(v)}, nil
}

func formatXLSFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func iso88591BytesToString(b []byte) string {
	r := make([]rune, len(b))
	for i, c := range b {
		r[i] = rune(c)
	}
	return string(r)
}

func cellsToDenseGrid(cells []sparseCell) ([][]string, error) {
	if len(cells) == 0 {
		return nil, nil
	}
	var maxRow, maxCol uint16
	for _, c := range cells {
		if c.row > maxRow {
			maxRow = c.row
		}
		if c.col > maxCol {
			maxCol = c.col
		}
	}
	nr := int(maxRow) + 1
	nc := int(maxCol) + 1
	if nr <= 0 || nc <= 0 {
		return nil, fmt.Errorf("xls: invalid grid dimensions")
	}
	grid := make([][]string, nr)
	for i := range grid {
		grid[i] = make([]string, nc)
	}
	seen := make([][]bool, nr)
	for i := range seen {
		seen[i] = make([]bool, nc)
	}
	for _, c := range cells {
		r, co := int(c.row), int(c.col)
		if seen[r][co] {
			return nil, fmt.Errorf("xls: duplicate cell at row %d col %d", r, co)
		}
		seen[r][co] = true
		grid[r][co] = c.text
	}
	for len(grid) > 0 {
		last := grid[len(grid)-1]
		empty := true
		for _, v := range last {
			if v != "" {
				empty = false
				break
			}
		}
		if !empty {
			break
		}
		grid = grid[:len(grid)-1]
	}
	return grid, nil
}

func cloneStringMatrix(in [][]string) [][]string {
	out := make([][]string, len(in))
	for i := range in {
		out[i] = append([]string(nil), in[i]...)
	}
	return out
}
