// +build ignore

package main

import (
	"compress/gzip"
	"log"
	"os"

	"github.com/bpineau/katafygio/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	header := &doc.GenManHeader{
		Title:   "KATAFYGIO",
		Section: "8",
		Source:  "Katafygio",
	}

	f, err := os.Create("katafygio.8.gz")
	if err != nil {
		log.Fatal(err)
	}

	zw := gzip.NewWriter(f)

	if err = doc.GenMan(cmd.RootCmd, header, zw); err != nil {
		log.Fatal(err)
	}

	if err = zw.Flush(); err != nil {
		log.Fatal(err)
	}

	if err = zw.Close(); err != nil {
		log.Fatal(err)
	}

	if err = f.Close(); err != nil {
		log.Fatal(err)
	}
}
