package xls

// BIFF record identifiers (16-bit little-endian record type on the wire).
// This package implements only a minimal linear subset used by [WriteXLS] / [ReadXLS], not full Excel 97 BIFF8.
//
// See Microsoft’s BIFF / XLS documentation (e.g. MS-XLS) for the full record taxonomy.
const (
	// BIFFRecordBOF marks the beginning of the BIFF stream (BOF / BIFF_VERSION-style opener; wire id 0x0809, bytes 09 08 LE).
	BIFFRecordBOF uint16 = 0x0809

	// BIFFRecordEOF terminates the linear stream (wire id 0x000A; payload length 0 in our writer).
	BIFFRecordEOF uint16 = 0x000A

	// BIFFRecordNumber is a NUMBER cell: row, column, XF index, then an IEEE-754 binary64 value.
	BIFFRecordNumber uint16 = 0x0203

	// BIFFRecordString is a string/label cell: row, column, XF index, byte length, then ISO-8859-1 payload.
	BIFFRecordString uint16 = 0x0204
)

// Exported aliases for cell record types (kept for stable API names).
const (
	// XLSIntType is the BIFF NUMBER record opcode (same as [BIFFRecordNumber]).
	XLSIntType = BIFFRecordNumber
	// XLSStringType is the BIFF string/label record opcode (same as [BIFFRecordString]).
	XLSStringType = BIFFRecordString
)

// BOF record body for the minimal opener written after the 4-byte BIFF header (type + length).
// Layout matches the historical “tiny .xls” writer: two version fields and two reserved words.
const (
	biffBOFRecordDataLen uint16 = 8    // value of the “record size” field for the BOF record payload
	biffBOFVersionMinor  uint16 = 0    // BIFF version / build (low word)
	biffBOFVersionMajor  uint16 = 0x10 // BIFF version / dialect (high word in this minimal layout)
	biffBOFReserved0     uint16 = 0
	biffBOFReserved1     uint16 = 0
)

// Cell record layout sizes (payload after the per-record 4-byte header).
const (
	biffNumberPayloadLen = 14 // row(2) + col(2) + xfIndex(2) + ieee754LittleEndian(8)

	// biffNumberValueOffset is the byte offset of the IEEE-754 value within a NUMBER record payload.
	biffNumberValueOffset = 6 // after row(2), col(2), xfIndex(2)

	biffStringHeaderBytes = 8 // row(2) + col(2) + xfIndex(2) + charCount(2) before ISO-8859-1 bytes

	biffDefaultXFIndex uint16 = 0 // style index; the linear writer always uses XF 0
)

// biffEOFRecordDataLen is the payload size for an EOF record (none).
const biffEOFRecordDataLen uint16 = 0

// OLE compound file binary signature (first four bytes of a CFB “.xls” container), not a linear BIFF stream.
var oleCFBHeaderPrefix = [...]byte{0xD0, 0xCF, 0x11, 0xE0}
