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

		lineno += 1

		doc := scanner.Text()

		go func(ctx context.Context, lineno int, doc string, cb TravelFunc, done_ch chan bool, err_ch chan error) {

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

			if err != nil {

				msg := fmt.Sprintf("[%d] %s", lineno, err)
				err_ch <- errors.New(msg)
				return
			}

		}(ctx, lineno, doc, cb, done_ch, err_ch)
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