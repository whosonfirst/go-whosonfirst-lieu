package main

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"flag"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"os"
	"sync"
)

func ParseFile(path string) error {

	log.Println("PARSE", path)

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

		go func(record string) {

			defer func() {
				wg.Done()
			}()

			select {
			case <-ctx.Done():
				return
			default:
				err := ParseRecord(record)

				if err != nil {
					log.Println(err)
					log.Println(record)
					cancel()
				}
			}

		}(record)

	}

	wg.Wait()

	return nil

}

func ParseRecord(raw string) error {

	var rsp gjson.Result

	rsp = gjson.Get(raw, "is_dupe")

	if !rsp.Exists() {
		return errors.New("Can't determine is_dupe")
	}

	is_dupe := rsp.Bool()

	rsp = gjson.Get(raw, "object.properties.wof:id")
	is_wof := rsp.Exists()

	rsp = gjson.Get(raw, "object.id")

	var id string

	if !rsp.Exists() {

		rsp = gjson.Get(raw, "object.properties.ref")

		if !rsp.Exists() {
			return errors.New("Can't determine object.id")
		}

		ref := rsp.String()

		h := sha1.New()
		io.WriteString(h, ref)

		rsp = gjson.Get(raw, "object.properties.@spider")

		if rsp.Exists() {
			spider := rsp.String()
			io.WriteString(h, spider)
		}

		dig := h.Sum(nil)
		enc := base64.RawURLEncoding.EncodeToString(dig)

		id = enc
		log.Println("ENC", enc)
	} else {

		id = rsp.String()
	}

	if is_dupe == false {

		if is_wof {
			return nil
		}

		// log.Println(id, "", "", "")
		return nil
	}

	rsp = gjson.Get(raw, "same_as")

	if !rsp.Exists() {
		return errors.New("Can't determine same_as")
	}

	return nil

	for _, o := range rsp.Array() {

		o_str := o.String()

		rsp = gjson.Get(o_str, "classification")
		classification := rsp.String()

		rsp = gjson.Get(o_str, "is_canonical")
		canonical := rsp.String()

		rsp = gjson.Get(o_str, "object.id")
		other_id := rsp.String()

		log.Println(id, classification, canonical, other_id)
	}

	return nil
}

func main() {

	flag.Parse()

	for _, path := range flag.Args() {

		err := ParseFile(path)

		if err != nil {
			break
		}
	}
}
