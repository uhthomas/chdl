package board

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func DetailChan8(u *url.URL) (board, thread string, err error) {
	s := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(s) > 0 {
		board = s[0]
	}
	if len(s) > 2 {
		thread = strings.Split(s[2], ".")[0]
	}
	return
}

type Chan8 struct {
	board string
}

func NewChan8(u *url.URL) (Chan8, error) {
	board, _, err := DetailChan8(u)
	return Chan8{board}, err
}

func (c8 Chan8) Board() string {
	return c8.board
}

func (c8 Chan8) Thread(thread string) Thread {
	return Chan8Thread{c8.board, thread}
}

func (c8 Chan8) Threads() (threads []Thread, err error) {
	for i := 0; i < 16; i++ {
		page, err := c8.Page(i + 1)
		if err != nil {
			return nil, err
		}

		if len(page) == 0 {
			break
		}

		threads = append(threads, page...)
	}
	return
}

func (c8 Chan8) Page(i int) (threads []Thread, err error) {
	res, err := http.Get(fmt.Sprintf("https://8ch.net/%s/%d.json", c8.board, i))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return
	}

	var data struct {
		Threads []struct {
			Posts []struct {
				Thread json.Number `json:"no,Number"`
			}
		}
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	for _, thread := range data.Threads {
		if len(thread.Posts) == 0 {
			continue
		}
		post := thread.Posts[0]
		threads = append(threads, Chan8Thread{c8.board, post.Thread.String()})
	}
	return
}

func (c8 Chan8) Posts() (posts []Post, err error) {
	threads, err := c8.Threads()
	if err != nil {
		return nil, err
	}

	for _, thread := range threads {
		tp, err := thread.Posts()
		if err != nil {
			return nil, err
		}
		posts = append(posts, tp...)
	}
	return
}

func (c8 Chan8) Files(excludeExtras bool) (files []File, err error) {
	threads, err := c8.Threads()
	if err != nil {
		return nil, err
	}

	for _, thread := range threads {
		tf, err := thread.Files(excludeExtras)
		if err != nil {
			return nil, err
		}
		files = append(files, tf...)
	}
	return
}

type Chan8Thread struct {
	board  string
	thread string
}

func (c8t Chan8Thread) URL() string {
	return fmt.Sprintf("https://8ch.net/%s/res/%s.json", c8t.board, c8t.thread)
}

func (c8t Chan8Thread) Board() string {
	return c8t.board
}

func (c8t Chan8Thread) Thread() string {
	return c8t.thread
}

func (c8t Chan8Thread) Posts() (posts []Post, err error) {
	res, err := http.Get(c8t.URL())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data struct {
		Posts []Chan8Post `json:"posts"`
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	for _, post := range data.Posts {
		post.board = c8t.board
		post.thread = c8t.thread
		posts = append(posts, post)
	}
	return
}

func (c8t Chan8Thread) Files(excludeExtras bool) (files []File, err error) {
	posts, err := c8t.Posts()
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		files = append(files, post.Files(excludeExtras)...)
	}
	return
}

type Chan8Post struct {
	board  string `json:",omitempty"`
	thread string `json:",omitempty"`

	Name      json.Number `json:"tim,Number,omitempty"`
	Extension string      `json:"ext,omitempty"`
	Extras    []struct {
		Name      string `json:"tim,Number"`
		Extension string `json:"ext"`
	} `json:"extra_files,omitempty"`
}

func (c8p Chan8Post) Board() string {
	return c8p.board
}

func (c8p Chan8Post) Files(excludeExtras bool) (files []File) {
	if c8p.Name != "" {
		files = append(files, Chan8File{c8p.board, c8p.thread, c8p.Name.String(), c8p.Extension[1:]})
	}
	if c8p.Extras == nil || excludeExtras {
		return
	}

	for _, extra := range c8p.Extras {
		files = append(files, Chan8File{c8p.board, c8p.thread, extra.Name, extra.Extension[1:]})
	}
	return
}

type Chan8File struct {
	board     string
	thread    string
	name      string
	extension string
}

func (c8f Chan8File) URL() string {
	return fmt.Sprintf("https://8ch.net/%s/src/%s.%s", c8f.board, c8f.name, c8f.extension)
}

func (c8f Chan8File) Board() string {
	return c8f.board
}

func (c8f Chan8File) Thread() string {
	return c8f.thread
}

func (c8f Chan8File) Name() string {
	return c8f.name
}

func (c8f Chan8File) Extension() string {
	return c8f.extension
}
