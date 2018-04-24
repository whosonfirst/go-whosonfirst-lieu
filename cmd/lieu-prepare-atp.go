package main

// WIP - not clear that this shouldn't be a generic ls-geojson pre-filter rather
// than anything ATP specific... (20171222/thisisaaronland)

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"github.com/openvenues/gopostal/parser"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-lieu"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

func Prepare(raw string, writer_ch chan []byte) error {

	// check for properties here...

	name := gjson.Get(raw, "properties.name")

	if !name.Exists() {
		return errors.New("missing name")
	}

	hn := gjson.Get(raw, "properties.addr:house_number")
	st := gjson.Get(raw, "properties.addr:street")

	if !hn.Exists() || !st.Exists() {

		// return errors.New(("record missing house_number and/or street")
		addr := gjson.Get(raw, "properties.addr:full")

		if !addr.Exists() {
			return errors.New("record missing address")
		}

		addr_full := addr.String()

		parsed := postal.ParseAddress(addr_full)

		house_number := ""
		street := ""

		for _, p := range parsed {

			if p.Label == "road" {
				street = p.Value
			}

			if p.Label == "house_number" {
				house_number = p.Value
			}
		}

		if house_number == "" {
			return errors.New("failed to parse house number from " + addr_full)
		}

		if street == "" {
			return errors.New("failed to parse street from " + addr_full)
		}

		var err error

		raw, err = sjson.Set(raw, "properties.addr:house_number", house_number)

		if err != nil {
			return err
		}

		raw, err = sjson.Set(raw, "properties.addr:street", street)

		if err != nil {
			return err
		}

	}

	body := []byte(raw)

	ok, err := lieu.HasRequiredProperties(body)

	if !ok {
		return err
	}

	body, err = lieu.EnstringifyProperties([]byte(raw))

	if err != nil {
		return err
	}

	var stub interface{}
	err = json.Unmarshal(body, &stub)

	if err != nil {
		return errors.New("failed to unmarshal")
	}

	body, err = json.Marshal(stub)

	if err != nil {
		return errors.New("failed to unmarshal")
	}

	writer_ch <- body
	return nil
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

	writer_ch := make(chan []byte)
	done_ch := make(chan bool)

	go func() {

		mu := new(sync.Mutex)

		for {
			select {
			case <-done_ch:
				return
			case body := <-writer_ch:
				mu.Lock()
				writer.Write(body)
				writer.Write([]byte("\n"))
				mu.Unlock()
			default:
				// pass
			}

		}

	}()

	var processed int64
	var skipped int64
	var nameless int64

	processed = 0
	skipped = 0
	nameless = 0

	for _, path := range flag.Args() {

		fh, err := os.Open(path)

		if err != nil {
			log.Fatal(err)
		}

		reader := bufio.NewReader(fh)
		wg := new(sync.WaitGroup)

		for {
			ln, err := reader.ReadString('\n')

			if err != nil {
				break
			}

			wg.Add(1)

			go func(ln string, writer_ch chan []byte, wg *sync.WaitGroup) {

				defer func() {
					wg.Done()
				}()

				atomic.AddInt64(&processed, 1)

				err := Prepare(ln, writer_ch)

				if err != nil {
					atomic.AddInt64(&skipped, 1)

					if err.Error() == "missing name" {
						atomic.AddInt64(&nameless, 1)
					}

					log.Println(err)
				}

			}(ln, writer_ch, wg)
		}

		wg.Wait()
	}

	done_ch <- true

	log.Printf("proccessed: %d skipped: %d nameless: %d\n", processed, skipped, nameless)
}
