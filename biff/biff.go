// Package biff holds wire constants for the minimal linear BIFF subset used by legacy .xls streams
// (not full Excel 97 BIFF8 and not OLE compound files). See MS-XLS and related specs for the full record taxonomy.
package biff

// Record opcodes (16-bit little-endian record type after stream offset).
const (
	// RecordBOF marks the beginning of the BIFF stream (BOF / BIFF_VERSION-style opener; wire id 0x0809, bytes 09 08 LE).
	RecordBOF uint16 = 0x0809

	// RecordEOF terminates the linear stream (wire id 0x000A; payload length 0 in our writer).
	RecordEOF uint16 = 0x000A

	// RecordNumber is a NUMBER cell: row, column, XF index, then an IEEE-754 binary64 value.
	RecordNumber uint16 = 0x0203

	// RecordString is a string/label cell: row, column, XF index, byte length, then ISO-8859-1 payload.
	RecordString uint16 = 0x0204
)

// XLSIntType is the BIFF NUMBER record opcode (same as [RecordNumber]).
const XLSIntType = RecordNumber

// XLSStringType is the BIFF string/label record opcode (same as [RecordString]).
const XLSStringType = RecordString

// BOF record body for the minimal opener written after the 4-byte BIFF header (type + length).
// Layout matches the historical “tiny .xls” writer: two version fields and two reserved words.
const (
	BOFPayloadLen   uint16 = 8    // value of the “record size” field for the BOF record payload
	BOFVersionMinor uint16 = 0    // BIFF version / build (low word)
	BOFVersionMajor uint16 = 0x10 // BIFF version / dialect (high word in this minimal layout)
	BOFReserved0    uint16 = 0
	BOFReserved1    uint16 = 0
)

// Cell record layout sizes (payload after the per-record 4-byte header).
const (
	NumberPayloadLen = 14 // row(2) + col(2) + xfIndex(2) + ieee754LittleEndian(8)

	// NumberValueOffset is the byte offset of the IEEE-754 value within a NUMBER record payload.
	NumberValueOffset = 6 // after row(2), col(2), xfIndex(2)

	// StringHeaderBytes is row(2) + col(2) + xfIndex(2) + charCount(2) before ISO-8859-1 bytes.
	StringHeaderBytes = 8

	// DefaultXFIndex is the style index; the linear writer always uses XF 0.
	DefaultXFIndex uint16 = 0
)

// EOFPayloadLen is the payload size for an EOF record (none).
const EOFPayloadLen uint16 = 0
