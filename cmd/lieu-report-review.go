package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"os"
	"sync"
)

type Row struct {
	Body []string
}

func NewRow(body []string) Row {

	r := Row{
		Body: body,
	}

	return r
}

func ParseFile(path string, writer_ch chan Row) error {

	// log.Println("PARSE", path)

	fh, err := os.Open(path)

	if err != nil {
		return err
	}

	defer fh.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := new(sync.WaitGroup)

	scanner := bufio.NewScanner(fh)

	for scanner.Scan() {

		wg.Add(1)

		record := scanner.Text()

		go func(record string, writer_ch chan Row) {

			defer func() {
				wg.Done()
			}()

			select {
			case <-ctx.Done():
				return
			default:
				err := ParseRecord(record, writer_ch)

				if err != nil {
					log.Println(err)
					log.Println(record)
					cancel()
				}
			}

		}(record, writer_ch)

	}

	wg.Wait()

	return nil

}

func ParseRecord(raw string, writer_ch chan Row) error {

	var rsp gjson.Result

	rsp = gjson.Get(raw, "is_dupe")

	if !rsp.Exists() {
		return errors.New("Can't determine is_dupe")
	}

	is_dupe := rsp.Bool()

	rsp = gjson.Get(raw, "object.properties.wof:id")
	is_wof := rsp.Exists()

	rsp = gjson.Get(raw, "object.id")

	if !rsp.Exists() {

		if !rsp.Exists() {
			return errors.New("Can't determine object.id")
		}
	}

	id := rsp.String()

	if is_dupe == false {

		if is_wof {
			return nil
		}

		out := []string{
			id,
			"unknown",
			"",
			"",
			"",
		}

		r := NewRow(out)

		writer_ch <- r
		return nil
	}

	rsp = gjson.Get(raw, "same_as")

	if !rsp.Exists() {
		return errors.New("Can't determine same_as")
	}

	for _, o := range rsp.Array() {

		o_str := o.String()

		rsp = gjson.Get(o_str, "classification")
		classification := rsp.String()

		rsp = gjson.Get(o_str, "is_canonical")
		canonical := rsp.String()

		rsp = gjson.Get(o_str, "object.id")
		other_id := rsp.String()

		rsp = gjson.Get(o_str, "object.properties.wof:id")
		other_is_wof := rsp.Exists()

		both_wof := ""

		if is_wof && other_is_wof {
			both_wof = "wof-wof"
		}

		out := []string{
			id,
			classification,
			canonical,
			other_id,
			both_wof,
		}

		r := NewRow(out)
		writer_ch <- r
	}

	return nil
}

func main() {

	var out = flag.String("out", "", "")
	flag.Parse()

	var fh io.WriteCloser

	if *out == "" {

		fh = os.Stdout
	} else {

		f, err := os.Create(*out)

		if err != nil {
			log.Fatal(err)
		}

		fh = f
	}

	writer := csv.NewWriter(fh)
	writer.Write([]string{"id", "classification", "is_canonical", "other_id", "is_wof"})

	writer_ch := make(chan Row)
	done_ch := make(chan bool)

	go func() {

		for {
			select {

			case <-done_ch:
				writer.Flush()
				fh.Close()
				return
			case row := <-writer_ch:
				writer.Write(row.Body)
			default:
				// pass
			}
		}
	}()

	for _, path := range flag.Args() {

		err := ParseFile(path, writer_ch)

		if err != nil {
			log.Println("failed to parse", path, err)
			break
		}
	}

	done_ch <- true
}
