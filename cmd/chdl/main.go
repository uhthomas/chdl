package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/uhthomas/chdl/board"
	"github.com/uhthomas/pipe"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	URL           = kingpin.Arg("url", "Board or thread URL").Required().URL()
	Limit         = kingpin.Flag("limit", "Concurrent download limit").Default("10").Short('l').Int()
	Output        = kingpin.Flag("out", "Output directory for downloaded files").Default("chdl").Short('o').String()
	ExcludeExtras = kingpin.Flag("exclude-extras", "Exclude extra files").Default("false").Short('e').Bool()
)

type Detail struct {
	File  board.File
	Error error
	Size  uint64
}

func main() {
	kingpin.Parse()

	_, thread, err := board.Detail(*URL)
	if err != nil {
		panic(err)
	}

	b, err := board.New(*URL)
	if err != nil {
		panic(err)
	}

	var (
		files []board.File
		ferr  error
	)

	if thread == "" {
		fmt.Printf("Are you sure you want to download everything in /%s? y/n: ", b.Board())
		var in string
		fmt.Scanln(&in)
		if in != "y" {
			return
		}
		files, ferr = b.Files(*ExcludeExtras)
	} else {
		files, ferr = b.Thread(thread).Files(*ExcludeExtras)
	}

	if ferr != nil {
		panic(ferr)
	}

	var (
		p          = pipe.New(*Limit)
		ch         = make(chan Detail)
		start      = time.Now()
		done       int
		downloaded uint64
	)

	for _, file := range files {
		go download(p, ch, file)
	}

	for {
		if done == len(files) {
			break
		}

		d := <-ch
		done++
		str := fmt.Sprintf("[%d/%d] ", done, len(files))
		if d.Error != nil {
			fmt.Printf("%sFailed to download %s/%s.%s\n\tReason: %s\n", str,
				d.File.Board(), d.File.Name(), d.File.Extension(), d.Error.Error())
		} else {
			fmt.Printf("%sFinished downloading %s/%s.%s (%s)\b\n", str,
				d.File.Board(), d.File.Name(), d.File.Extension(), humanize.Bytes(d.Size))
			downloaded += d.Size
		}
	}

	fmt.Printf("Downloaded %d files (%s) in %s\n", len(files), humanize.Bytes(downloaded), time.Since(start))
}

func download(p pipe.Pipe, ch chan Detail, f board.File) {
	p.Increment()
	go func() {
		defer p.Decrement()

		if err := os.MkdirAll(filepath.Join(*Output, f.Board()), os.ModePerm); err != nil {
			ch <- Detail{File: f, Error: err}
			return
		}

		if _, err := os.Stat(dir(f)); err == nil {
			ch <- Detail{File: f, Error: errors.New("File already exists")}
			return
		}

		res, err := http.Get(f.URL())
		if err != nil {
			ch <- Detail{File: f, Error: err}
			return
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			ch <- Detail{File: f, Error: fmt.Errorf("unexpected status code %d", res.StatusCode)}
			return
		}

		out, err := os.Create(dir(f))
		if err != nil {
			ch <- Detail{File: f, Error: err}
			return
		}
		defer out.Close()

		w, err := io.Copy(out, res.Body)
		ch <- Detail{f, err, uint64(w)}
	}()
}

func dir(f board.File) string {
	return filepath.Join(*Output, f.Board(), fmt.Sprintf("%s.%s", f.Name(),
		f.Extension()))
}
