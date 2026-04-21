package xls

// Export format constants (aligned with Mapbender FOM ExportResponse).
const (
	TypeCSV  = "csv"
	TypeXLS  = "xls"
	TypeXLSX = "xlsx"
)

// BIFF cell record types (Mapbender / legacy Excel writer).
const (
	XLSIntType    = 0x203
	XLSStringType = 0x204
)
