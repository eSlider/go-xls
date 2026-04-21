package xls

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// MIME-style content types for tabular downloads.
const (
	ContentTypeCSV = "text/csv;charset=UTF-16LE"
	ContentTypeXLS = "application/vnd.ms-excel"
)

// EnableDownloadHeaders sets Cache-Control, Pragma, and Expires for attachment responses.
func EnableDownloadHeaders(h http.Header) {
	h.Set("Cache-Control", "private")
	h.Set("Pragma", "no-cache")
	h.Set("Expires", "0")
}

// DisableDownloadHeaders removes cache headers added by [EnableDownloadHeaders].
func DisableDownloadHeaders(h http.Header) {
	h.Del("Cache-Control")
	h.Del("Pragma")
	h.Del("Expires")
}

// WriteAttachment sets Content-Type, Content-Disposition, optional Content-Length,
// optional download cache headers, then copies body into w.
//
// If size >= 0, Content-Length is set to size. If size < 0, Content-Length is omitted (chunked transfer).
func WriteAttachment(w http.ResponseWriter, fileName, contentType string, body io.Reader, size int64, enableDownload bool) error {
	if enableDownload {
		EnableDownloadHeaders(w.Header())
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	if size >= 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	}
	w.WriteHeader(http.StatusOK)
	_, err := io.Copy(w, body)
	return err
}
