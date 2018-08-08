package main

import (
	"encoding/json"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-lieu/travel"
	"io"
	"log"
	"os"
	"strings"
)

func ensureValidJSON(doc string) error {

	var err error
	var stub interface{}

	dec := json.NewDecoder(strings.NewReader(doc))

	for {

		err = dec.Decode(&stub)

		if err != nil {
			break
		}
	}

	if err != io.EOF {
		return err
	}

	return nil
}

func main() {

	flag.Parse()

	for _, path := range flag.Args() {

		err := travel.Travel(path, ensureValidJSON)

		if err != nil {
			log.Fatal(path, err)
		}
	}

	os.Exit(0)
}
