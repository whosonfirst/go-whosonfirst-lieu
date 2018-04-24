package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/go-whosonfirst-lieu"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	var mode = flag.String("mode", "repo", "")
	var out = flag.String("out", "", "")
	var timings = flag.Bool("timings", false, "")
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

	var counter int64
	counter = 0

	writer_ch := make(chan []byte)
	done_ch := make(chan bool)
	counter_ch := make(chan bool)

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		defer func() {
			counter_ch <- true
		}()

		path, err := index.PathForContext(ctx)

		if err != nil {
			return err
		}

		ok, err := utils.IsPrincipalWOFRecord(fh, ctx)

		if err != nil {
			log.Println("Failed to determine if principal", path)
			return err
		}

		if !ok {
			return nil
		}

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil {
			log.Println("Failed to load feature", path)
			return err
		}

		if f.Placetype() != "venue" {
			return nil
		}

		d, err := whosonfirst.IsDeprecated(f)

		if err != nil {
			log.Println("Failed to determine deprecated", path)
			return err
		}

		if d.IsTrue() && d.IsKnown() {
			return nil
		}

		s, err := whosonfirst.IsSuperseded(f)

		if err != nil {
			log.Println("Failed to determine if superseded", path)
			return err
		}

		if s.IsTrue() && s.IsKnown() {
			return nil
		}

		body, err := lieu.PrepareFeature(f.Bytes())

		if err != nil {
			return err
		}

		// there should be a better way... (20171222/thisisaaronland)

		var stub interface{}
		err = json.Unmarshal(body, &stub)

		if err != nil {
			log.Println("Failed to unmarshal", path)
			return err
		}

		body, err = json.Marshal(stub)

		if err != nil {
			log.Println("Failed to marshal", path)
			return err
		}

		writer_ch <- body
		return nil
	}

	t1 := time.Now()

	go func() {

		mu := new(sync.Mutex)

		for {

			select {
			case <-counter_ch:
				atomic.AddInt64(&counter, 1)
			case <-done_ch:
				writer.Close()
				break
			case body := <-writer_ch:
				mu.Lock()
				writer.Write(body)
				writer.Write([]byte("\n"))
				mu.Unlock()
			}
		}
	}()

	if *timings {

		go func() {
			for {

				select {
				case <-time.After(30 * time.Second):
					c := atomic.LoadInt64(&counter)
					t2 := time.Since(t1)
					log.Printf("time to process %d records %v\n", c, t2)
				}

			}
		}()
	}

	i, err := index.NewIndexer(*mode, cb)

	if err != nil {
		log.Fatal(err)
	}

	for _, path := range flag.Args() {

		ta := time.Now()

		err := i.IndexPath(path)

		if err != nil {
			log.Fatal(err)
		}

		tb := time.Since(ta)

		if *timings {
			log.Printf("time to prepare %s %v\n", path, tb)
		}

	}

	t2 := time.Since(t1)

	if *timings {
		c := atomic.LoadInt64(&counter)
		log.Printf("time to prepare all %d records %v\n", c, t2)
	}

	done_ch <- true

}
