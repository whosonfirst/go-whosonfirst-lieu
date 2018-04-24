package main

import (
	"flag"
	"github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-lieu"
	"log"
	"os"
)

func main() {

	var debug = flag.Bool("debug", false, "")

	flag.Parse()

	for _, path := range flag.Args() {

		fh, err := os.Open(path)

		if err != nil {
			log.Fatal(err)
		}

		if *debug {

			err = lieu.Prepare(fh, os.Stdout)

			if err != nil {
				log.Fatal(err)
			}

		} else {

			out, err := atomicfile.New(path, 0644)

			if err != nil {
				log.Fatal(err)
			}

			err = lieu.Prepare(fh, out)

			if err != nil {

				err = out.Abort()

				if err != nil {
					log.Fatal(err)
				}
			}

			err = out.Close()

			if err != nil {
				log.Fatal(err)
			}
		}
	}

}
