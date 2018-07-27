package main

import (
	"bufio"
	"compress/bzip2"
	"flag"
	"github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-lieu"
	"io"
	"log"
	"os"
)

func main() {

	var bunzip = flag.Bool("bunzip", false, "")
	var debug = flag.Bool("debug", false, "")

	flag.Parse()

	for _, path := range flag.Args() {

		var r io.Reader

		fh, err := os.Open(path)

		if err != nil {
			log.Fatal(err)
		}

		r = fh

		if *bunzip {

			br := bufio.NewReader(fh)
			r = bzip2.NewReader(br)
		}

		if *debug {

			err = lieu.Prepare(r, os.Stdout)

			if err != nil {
				log.Fatal(err)
			}

		} else {

			// strip path here...

			out, err := atomicfile.New(path, 0644)

			if err != nil {
				log.Fatal(err)
			}

			err = lieu.Prepare(r, out)

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
