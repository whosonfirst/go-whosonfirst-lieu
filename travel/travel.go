package travel

import (
	"bufio"
	"compress/bzip2"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type TravelFunc func(string) error

func Travel(path string, cb TravelFunc) error {

	t1 := time.Now()

	defer func() {
		log.Printf("time to travel %s %v\n", path, time.Since(t1))
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

	for scanner.Scan() {

		doc := scanner.Text()
		err := cb(doc)

		if err != nil {
		   return err
		}
	}	

	return nil

	// THIS LEAKS MEMORY... SOMEWHERE BUT WHERE?
	// (20180808/thisisaaronland)

	done_ch := make(chan bool)
	err_ch := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	count_throttles := 10
	throttle_ch := make(chan bool, count_throttles)

	for i := 0; i < count_throttles; i++ {
	    throttle_ch <- true
	}

	for scanner.Scan() {

		<- throttle_ch

		lineno += 1

		doc := scanner.Text()

		go func(lineno int, doc string) {

			defer func() {
				done_ch <- true
			}()

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			err := cb(doc)

			throttle_ch <- true

			if err != nil {
				msg := fmt.Sprintf("[%d] %s", lineno, err)
				err_ch <- errors.New(msg)
			}

		}(lineno, doc)
	}

	remaining := lineno

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case e := <-err_ch:
			log.Println(e)
		default:
			// pass
		}
	}

	return nil
}