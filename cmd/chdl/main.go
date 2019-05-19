package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/uhthomas/chdl/pkg"
	"github.com/uhthomas/pipe"
)

func download(ctx context.Context, url, dir, name string) (n int64, err error) {
	if err := os.MkdirAll(dir, 0777); err != nil {
		return 0, err
	}

	if _, err := os.Stat(filepath.Join(dir, name)); os.IsExist(err) {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	res, err := (&http.Client{Timeout: 10 * time.Second}).Do(req.WithContext(ctx))
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpect status code %d", res.StatusCode)
	}

	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return io.Copy(f, res.Body)
}

func main() {
	var (
		limit         = flag.Int("limit", 10, "Concurrent download limit")
		out           = flag.String("out", "chdl", "Output directory for files.")
		excludeExtras = flag.Bool("exclude-extras", false, "Don't download extra files")
	)
	flag.Parse()

	u, err := url.Parse(flag.Arg(0))
	if err != nil {
		log.Fatal("invalid chdl or thread url")
	}

	_, thread, err := pkg.Detail(u)
	if err != nil {
		panic(err)
	}

	b, err := pkg.New(u)
	if err != nil {
		panic(err)
	}

	var files []pkg.File

	if thread == "" {
		fmt.Printf("Are you sure you want to download everything in /%s? y/n: ", b.Board())
		var in string
		if _, err := fmt.Scanln(&in); err != nil {
			log.Fatal(err)
		}
		if in != "y" {
			return
		}
		files, err = b.Files(*excludeExtras)
	} else {
		files, err = b.Thread(thread).Files(*excludeExtras)
	}

	if err != nil {
		log.Fatal(err)
	}

	var (
		ctx   = context.Background()
		delta uint64
		done  uint64
		start = time.Now()
		p     = pipe.New(*limit)
		wg    sync.WaitGroup
	)

	wg.Add(len(files))
	for _, f := range files {
		f := f

		p.Increment()
		go func() {
			defer wg.Done()
			defer p.Decrement()

			var (
				dir  = filepath.Join(*out, f.Board())
				name = fmt.Sprintf("%s.%s", f.Name(), f.Extension())
			)

			n, err := download(ctx, f.URL(), dir, name)

			prefix := fmt.Sprintf("[%d/%d]", atomic.AddUint64(&done, 1), len(files))

			if err != nil {
				fmt.Printf(
					"%s Failed to download %s\nReason: %s\n",
					prefix,
					filepath.Join(f.Board(), name),
					err.Error(),
				)
				return
			}

			fmt.Printf(
				"%s Downloaded %s (%s)\n",
				prefix,
				filepath.Join(f.Board(), name),
				humanize.Bytes(uint64(n)),
			)

			atomic.AddUint64(&delta, uint64(n))
		}()
	}

	wg.Wait()

	fmt.Printf(
		"Downloaded %d files (%s) in %s\n",
		len(files),
		humanize.Bytes(delta),
		time.Since(start),
	)
}
