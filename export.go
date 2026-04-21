package xls

// Format identifiers for helpers and documentation.
const (
	TypeCSV  = "csv"
	TypeXLS  = "xls"
	TypeXLSX = "xlsx"
)

// BIFF legacy cell record types (binary .xls subset).
const (
	XLSIntType    = 0x203
	XLSStringType = 0x204
)
