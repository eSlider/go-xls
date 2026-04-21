package xls

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"
	"strings"
)

// GenXLS generates a legacy .xls (BIFF) workbook matching Mapbender ExportResponse::genXLS
// layout on little-endian hosts (PHP pack "s" / "d").
func GenXLS(tab Table, detectHead bool) ([]byte, error) {
	tab = normalizeTable(tab)

	var buf bytes.Buffer

	// Excel BOF (BIFF2/BIFF5-style minimal writer used by Mapbender).
	writeU16LE(&buf, 0x809)
	writeU16LE(&buf, 0x8)
	writeU16LE(&buf, 0)
	writeU16LE(&buf, 0x10)
	writeU16LE(&buf, 0)
	writeU16LE(&buf, 0)

	rowNum := uint16(0)

	if detectHead && len(tab.Rows) > 0 && len(tab.Columns) > 0 && !allPHPNumericKeys(tab.Columns) {
		colNum := uint16(0)
		for _, key := range tab.Columns {
			cell, err := encodeXLSStringCell(rowNum, colNum, key)
			if err != nil {
				return nil, err
			}
			buf.Write(cell)
			colNum++
		}
		rowNum++
	}

	for _, row := range tab.Rows {
		colNum := uint16(0)
		for range tab.Columns {
			val := ""
			if int(colNum) < len(row) {
				val = row[colNum]
			}
			cell, err := encodeXLSCell(rowNum, colNum, val)
			if err != nil {
				return nil, err
			}
			buf.Write(cell)
			colNum++
		}
		rowNum++
	}

	// Excel EOF
	writeU16LE(&buf, 0x0A)
	writeU16LE(&buf, 0)

	return buf.Bytes(), nil
}

func encodeXLSCell(row, col uint16, raw string) ([]byte, error) {
	s := strings.TrimSpace(raw)
	if phpIsNumeric(s) {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return encodeXLSStringCell(row, col, raw)
		}
		if math.IsInf(f, 0) || math.IsNaN(f) {
			return encodeXLSStringCell(row, col, raw)
		}
		return encodeXLSNumberCell(row, col, f), nil
	}
	return encodeXLSStringCell(row, col, raw)
}

func encodeXLSNumberCell(row, col uint16, v float64) []byte {
	var buf bytes.Buffer
	writeU16LE(&buf, XLSIntType)
	writeU16LE(&buf, 14)
	writeU16LE(&buf, row)
	writeU16LE(&buf, col)
	writeU16LE(&buf, 0)
	_ = binary.Write(&buf, binary.LittleEndian, v)
	return buf.Bytes()
}

func encodeXLSStringCell(row, col uint16, s string) ([]byte, error) {
	b, err := toISO88591(s)
	if err != nil {
		return nil, err
	}
	l := len(b)
	var buf bytes.Buffer
	writeU16LE(&buf, XLSStringType)
	writeU16LE(&buf, uint16(8+l))
	writeU16LE(&buf, row)
	writeU16LE(&buf, col)
	writeU16LE(&buf, 0)
	writeU16LE(&buf, uint16(l))
	buf.Write(b)
	return buf.Bytes(), nil
}

func writeU16LE(w *bytes.Buffer, v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	_, _ = w.Write(b[:])
}

func allPHPNumericKeys(keys []string) bool {
	for _, k := range keys {
		if !phpIsNumeric(k) {
			return false
		}
	}
	return true
}

// phpIsNumeric approximates PHP is_numeric for array keys used by Mapbender.
func phpIsNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	// PHP treats leading +/- and floats, hex in older PHP - stick to ParseFloat parity for decimals.
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func toISO88591(s string) ([]byte, error) {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		if r <= 0xff {
			out = append(out, byte(r))
			continue
		}
		out = append(out, '?')
	}
	return out, nil
}
