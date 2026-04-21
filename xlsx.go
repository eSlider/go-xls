package xls

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// GenXLSX builds an .xlsx workbook (OpenXML) from tabular data, analogous to Mapbender's SimpleXLSXGen::fromArray.
func GenXLSX(tab Table, detectHead bool) ([]byte, error) {
	tab = normalizeTable(tab)

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, fmt.Errorf("xls: excelize default sheet missing")
	}

	r := 1
	if detectHead && len(tab.Rows) > 0 && len(tab.Columns) > 0 && !allPHPNumericKeys(tab.Columns) {
		for c, name := range tab.Columns {
			cell, err := excelize.CoordinatesToCellName(c+1, r)
			if err != nil {
				return nil, err
			}
			if err := f.SetCellStr(sheet, cell, name); err != nil {
				return nil, err
			}
		}
		r++
	}

	for _, row := range tab.Rows {
		for c := range tab.Columns {
			cell, err := excelize.CoordinatesToCellName(c+1, r)
			if err != nil {
				return nil, err
			}
			val := ""
			if c < len(row) {
				val = row[c]
			}
			if err := setCellSmart(f, sheet, cell, val); err != nil {
				return nil, err
			}
		}
		r++
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func setCellSmart(f *excelize.File, sheet, cell, val string) error {
	s := strings.TrimSpace(val)
	if s == "" {
		return f.SetCellValue(sheet, cell, "")
	}
	if phpIsNumeric(s) {
		fl, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f.SetCellFloat(sheet, cell, fl, -1, 64)
		}
	}
	return f.SetCellStr(sheet, cell, val)
}
