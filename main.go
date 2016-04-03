package main

import (
	"fmt"
	"go-chdl/thread"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/6f7262/pipe"
	humanize "github.com/dustin/go-humanize"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	URL           = kingpin.Arg("url", "Thread URL").Required().URL()
	Limit         = kingpin.Flag("limit", "Concurrent download limit").Default("10").Short('l').Int()
	Output        = kingpin.Flag("out", "Output directory for downloaded files").Default("chdl").Short('o').String()
	IncludeExtras = kingpin.Flag("exclude-extras", "Exclude extra files").Default("false").Short('e').Bool()
)

type Detail struct {
	File  thread.File
	Error error
	Size  uint64
}

func main() {
	kingpin.Parse()

	if err := os.MkdirAll(*Output, os.ModePerm); err != nil {
		panic(err)
	}

	t, err := thread.New(*URL)
	if err != nil {
		panic(err)
	}

	files, err := t.Files(*IncludeExtras)
	if err != nil {
		panic(err)
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
		d := <-ch
		done++
		str := fmt.Sprintf("[%d/%d] ", done, len(files))
		if d.Error != nil {
			fmt.Printf("%sFailed to download %s/%s.%s\n\tReason: %s\n", str,
				d.File.Board(), d.File.Name(), d.File.Extension(), d.Error.Error())
			continue
		}
		fmt.Printf("%sFinished downloading %s/%s.%s (%s)\b\n", str,
			d.File.Board(), d.File.Name(), d.File.Extension(), humanize.Bytes(d.Size))

		downloaded += d.Size

		if done == len(files) {
			break
		}
	}

	fmt.Printf("Downloaded %d files (%s) in %s\n", len(files), humanize.Bytes(downloaded), time.Since(start))
}

func download(p pipe.Pipe, ch chan Detail, file thread.File) {
	p.Increment()
	go func() {
		defer p.Decrement()

		res, err := http.Get(file.URL())
		if err != nil {
			ch <- Detail{File: file, Error: err}
			return
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			ch <- Detail{File: file, Error: fmt.Errorf("unexpected status code %d", res.StatusCode)}
			return
		}

		out, err := os.Create(filepath.Join(*Output, fmt.Sprintf("%s.%s", file.Name(), file.Extension())))
		if err != nil {
			ch <- Detail{File: file, Error: err}
			return
		}
		defer out.Close()

		w, err := io.Copy(out, res.Body)
		ch <- Detail{File: file, Error: err, Size: uint64(w)}
	}()
}
