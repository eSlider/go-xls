// Command genxls writes a minimal linear BIFF .xls you can open locally (e.g. LibreOffice, older Excel).
package main

import (
	"flag"
	"log"
	"os"

	"github.com/eslider/go-xls/v2"
)

func main() {
	out := flag.String("o", "sample.xls", "output .xls path")
	flag.Parse()

	tab := xls.Table{
		Columns: []string{"name", "qty", "note"},
		Rows: [][]string{
			{"apple", "12", "in stock"},
			{"banana", "3", "ripe"},
			{"cherry", "0", "sold out"},
		},
	}
	f, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := xls.WriteXLS(f, tab, true); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	fi, err := os.Stat(*out)
	if err != nil {
		log.Fatalf("wrote %s", *out)
	}
	log.Printf("wrote %s (%d bytes)", *out, fi.Size())
}
