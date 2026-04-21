package xls

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"strconv"
	"strings"
)

// WriteXLS writes a legacy binary .xls workbook as a linear little-endian BIFF record stream
// ([BIFFRecordBOF], [BIFFRecordString], [BIFFRecordNumber], [BIFFRecordEOF]). It is not an OLE compound file.
func WriteXLS(w io.Writer, tab Table, detectHead bool) error {
	tab = normalizeTable(tab)

	writeU16 := func(v uint16) error {
		var b [2]byte
		binary.LittleEndian.PutUint16(b[:], v)
		_, err := w.Write(b[:])
		return err
	}

	// BOF: record type + declared payload length, then fixed minimal version words (see xls_biff.go).
	for _, v := range []uint16{
		BIFFRecordBOF,
		biffBOFRecordDataLen,
		biffBOFVersionMinor,
		biffBOFVersionMajor,
		biffBOFReserved0,
		biffBOFReserved1,
	} {
		if err := writeU16(v); err != nil {
			return err
		}
	}

	rowNum := uint16(0)

	if detectHead && len(tab.Rows) > 0 && len(tab.Columns) > 0 && !allKeysNumeric(tab.Columns) {
		colNum := uint16(0)
		for _, key := range tab.Columns {
			cell, err := encodeXLSStringCell(rowNum, colNum, key)
			if err != nil {
				return err
			}
			if _, err := w.Write(cell); err != nil {
				return err
			}
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
				return err
			}
			if _, err := w.Write(cell); err != nil {
				return err
			}
			colNum++
		}
		rowNum++
	}

	if err := writeU16(BIFFRecordEOF); err != nil {
		return err
	}
	return writeU16(biffEOFRecordDataLen)
}

func encodeXLSCell(row, col uint16, raw string) ([]byte, error) {
	s := strings.TrimSpace(raw)
	if isNumericString(s) {
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
	writeU16LE(&buf, BIFFRecordNumber)
	writeU16LE(&buf, biffNumberPayloadLen)
	writeU16LE(&buf, row)
	writeU16LE(&buf, col)
	writeU16LE(&buf, biffDefaultXFIndex)
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
	writeU16LE(&buf, BIFFRecordString)
	writeU16LE(&buf, uint16(biffStringHeaderBytes+l))
	writeU16LE(&buf, row)
	writeU16LE(&buf, col)
	writeU16LE(&buf, biffDefaultXFIndex)
	writeU16LE(&buf, uint16(l))
	buf.Write(b)
	return buf.Bytes(), nil
}

func writeU16LE(w *bytes.Buffer, v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	_, _ = w.Write(b[:])
}

// allKeysNumeric reports whether every column name parses as a decimal number (no header row in that case).
func allKeysNumeric(keys []string) bool {
	for _, k := range keys {
		if !isNumericString(k) {
			return false
		}
	}
	return true
}

// isNumericString reports whether s looks like a decimal number (after trim).
func isNumericString(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
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
