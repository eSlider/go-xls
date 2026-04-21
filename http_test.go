package xls

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteAttachment(t *testing.T) {
	rec := httptest.NewRecorder()
	body := strings.NewReader("abc")
	if err := WriteAttachment(rec, "export.xls", ContentTypeXLS, body, 3, true); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	defer res.Body.Close()
	if res.Header.Get("Content-Type") != ContentTypeXLS {
		t.Fatalf("content-type=%q", res.Header.Get("Content-Type"))
	}
	if !strings.Contains(res.Header.Get("Content-Disposition"), `filename="export.xls"`) {
		t.Fatalf("disposition=%q", res.Header.Get("Content-Disposition"))
	}
	if res.Header.Get("Cache-Control") != "private" {
		t.Fatalf("cache-control=%q", res.Header.Get("Cache-Control"))
	}
	out, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "abc" {
		t.Fatalf("body=%q", out)
	}
}
