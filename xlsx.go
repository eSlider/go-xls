package xls

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// WriteXLSX writes a single-sheet OpenXML .xlsx workbook to w.
func WriteXLSX(w io.Writer, tab Table, detectHead bool) error {
	tab = normalizeTable(tab)

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	if sheet == "" {
		return fmt.Errorf("xls: excelize default sheet missing")
	}

	r := 1
	if detectHead && len(tab.Rows) > 0 && len(tab.Columns) > 0 && !allKeysNumeric(tab.Columns) {
		for c, name := range tab.Columns {
			cell, err := excelize.CoordinatesToCellName(c+1, r)
			if err != nil {
				return err
			}
			if err := f.SetCellStr(sheet, cell, name); err != nil {
				return err
			}
		}
		r++
	}

	for _, row := range tab.Rows {
		for c := range tab.Columns {
			cell, err := excelize.CoordinatesToCellName(c+1, r)
			if err != nil {
				return err
			}
			val := ""
			if c < len(row) {
				val = row[c]
			}
			if err := setCellSmart(f, sheet, cell, val); err != nil {
				return err
			}
		}
		r++
	}

	return f.Write(w)
}

func setCellSmart(f *excelize.File, sheet, cell, val string) error {
	s := strings.TrimSpace(val)
	if s == "" {
		return f.SetCellValue(sheet, cell, "")
	}
	if isNumericString(s) {
		fl, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f.SetCellFloat(sheet, cell, fl, -1, 64)
		}
	}
	return f.SetCellStr(sheet, cell, val)
}
