package main

import (
	"bufio"
	"encoding/json"
	"flag"
	// "github.com/openvenues/gopostal/parser"
	// "github.com/tidwall/gjson"
	// "github.com/tidwall/sjson"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

func Prepare(raw string, writer_ch chan []byte) {

	// check for properties here...

	// name
	// addr:housenumber
	// addr:street
	// add:full

	var stub interface{}
	err := json.Unmarshal([]byte(raw), &stub)

	if err != nil {
		log.Println("failed to unmarshal")
		return
	}

	body, err := json.Marshal(stub)

	if err != nil {
		log.Println("failed to unmarshal")
		return
	}

	writer_ch <- body
}

func main() {

	var out = flag.String("out", "", "")
	// var timings = flag.Bool("timings", false, "")
	var procs = flag.Int("processes", runtime.NumCPU()*2, "")

	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	var writer io.WriteCloser

	if *out != "" {
		fh, err := os.Create(*out)

		if err != nil {
			log.Fatal(err)
		}

		writer = fh
	} else {
		writer = os.Stdout
	}

	for _, path := range flag.Args() {

		fh, err := os.Open(path)

		if err != nil {
			log.Fatal(err)
		}

		writer_ch := make(chan []byte)
		done_ch := make(chan bool)

		mu := new(sync.Mutex)

		scanner := bufio.NewScanner(fh)

		var remaining int64
		remaining = 0

		for scanner.Scan() {
			ln := scanner.Text()
			atomic.AddInt64(&remaining, 1)
			go Prepare(ln, writer_ch)
		}

		for {
			select {
			case <-done_ch:
				break
			case body := <-writer_ch:
				mu.Lock()
				writer.Write(body)
				writer.Write([]byte("\n"))
				mu.Unlock()

				i := atomic.AddInt64(&remaining, -1)

				if i <= 0 {
					break
				}
			}

		}

	}
}
