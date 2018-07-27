package main

import (
	"bufio"
	"compress/bzip2"
	"context"
	"encoding/json"
	"flag"
	"github.com/tidwall/gjson"
	_ "github.com/whosonfirst/go-whosonfirst-lieu"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func ValidateFile(path string, throttle_ch chan bool) error {

	t1 := time.Now()

	defer func() {
		log.Printf("time to validate %s %v\n", path, time.Since(t1))
	}()

	var r io.Reader

	fh, err := os.Open(path)

	if err != nil {
		return err
	}

	r = fh

	if filepath.Ext(path) == ".bz2" {
		br := bufio.NewReader(fh)
		r = bzip2.NewReader(br)
	}

	scanner := bufio.NewScanner(r)
	lineno := 0

	done_ch := make(chan bool)
	err_ch := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for scanner.Scan() {

		// <- throttle_ch
		lineno += 1

		doc := scanner.Text()

		go validateDocument(ctx, doc, lineno, done_ch, err_ch)
	}

	remaining := lineno

	for remaining > 0 {

		select {
		case <-done_ch:
			// throttle_ch <- true
			remaining -= 1
		case e := <-err_ch:
			log.Println(e)
		default:
			// pass
		}
	}

	return nil
}

func validateDocument(ctx context.Context, doc string, lineno int, done_ch chan bool, err_ch chan error) {

	defer func() {
		done_ch <- true
	}()

	select {
	case <-ctx.Done():
		return
	default:
		// pass
	}

	err := ensureValidJSON(doc)

	if err != nil {
		err_ch <- err
		return
	}

	rsp := gjson.Get(doc, "properties.addr:phone")

	if rsp.Exists() {
		log.Printf("PHONE '%s'\n", rsp.String())
	}
}

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

	var procs = flag.Int("processes", runtime.NumCPU()*2, "The number of concurrent processes to use")
	// var strict = flag.Bool("strict", false, "Whether or not to trigger a fatal error when invalid JSON is encountered")
	// var stats = flag.Bool("stats", false, "Be chatty, with counts and stuff")

	flag.Parse()

	throttle_ch := make(chan bool, *procs)

	for i := 0; i < *procs; i++ {
		throttle_ch <- true
	}

	for _, path := range flag.Args() {
		ValidateFile(path, throttle_ch)
	}

	os.Exit(0)
}
