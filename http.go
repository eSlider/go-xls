package xls

import (
	"fmt"
	"net/http"
	"strconv"
)

// Content types aligned with Mapbender ExportResponse::setType.
const (
	ContentTypeCSV  = "text/csv;charset=UTF-16LE"
	ContentTypeXLS  = "application/vnd.ms-excel"
	ContentTypeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
)

// EnableDownloadHeaders sets Cache-Control / Pragma / Expires like Mapbender enableDownload().
func EnableDownloadHeaders(h http.Header) {
	h.Set("Cache-Control", "private")
	h.Set("Pragma", "no-cache")
	h.Set("Expires", "0")
}

// DisableDownloadHeaders removes the download-related cache headers.
func DisableDownloadHeaders(h http.Header) {
	h.Del("Cache-Control")
	h.Del("Pragma")
	h.Del("Expires")
}

// WriteAttachment writes body with Content-Type, Content-Disposition, Content-Length,
// and optional download headers (Mapbender ExportResponse behavior).
func WriteAttachment(w http.ResponseWriter, fileName, contentType string, body []byte, enableDownload bool) {
	if enableDownload {
		EnableDownloadHeaders(w.Header())
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
