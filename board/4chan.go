package board

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Chan4 struct {
	board string
}

func NewChan4(u *url.URL) (Chan4, error) {
	board, _, err := DetailChan4(u)
	return Chan4{board}, err
}

func DetailChan4(u *url.URL) (board, thread string, err error) {
	s := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(s) > 0 {
		board = s[0]
	}
	if len(s) > 2 {
		thread = s[2]
	}
	return
}

func (c4 Chan4) Board() string {
	return c4.board
}

func (c4 Chan4) Thread(thread string) Thread {
	return Chan4Thread{c4.board, thread}
}

func (c4 Chan4) Threads() (threads []Thread, err error) {
	for i := 0; i < 10; i++ {
		page, err := c4.Page(i + 1)
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

func (c4 Chan4) Page(i int) (threads []Thread, err error) {
	res, err := http.Get(fmt.Sprintf("https://a.4cdn.org/%s/%d.json", c4.board, i))
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
		threads = append(threads, Chan4Thread{c4.board, post.Thread.String()})
	}
	return
}

func (c4 Chan4) Posts() (posts []Post, err error) {
	threads, err := c4.Threads()
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

func (c4 Chan4) Files(excludeExtras bool) (files []File, err error) {
	threads, err := c4.Threads()
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

type Chan4Thread struct {
	board  string
	thread string
}

func (c4t Chan4Thread) URL() string {
	return fmt.Sprintf("https://a.4cdn.org/%s/thread/%s.json", c4t.board, c4t.thread)
}

func (c4t Chan4Thread) Board() string {
	return c4t.board
}

func (c4t Chan4Thread) Thread() string {
	return c4t.thread
}

func (c4t Chan4Thread) Posts() (posts []Post, err error) {
	res, err := http.Get(c4t.URL())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data struct {
		Posts []Chan4Post `json:"posts"`
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	for _, post := range data.Posts {
		post.board = c4t.board
		post.thread = c4t.thread
		posts = append(posts, post)
	}
	return
}

func (c4t Chan4Thread) Files(excludeExtras bool) (files []File, err error) {
	posts, err := c4t.Posts()
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		files = append(files, post.Files(excludeExtras)...)
	}
	return
}

type Chan4Post struct {
	board  string `json:",omitempty"`
	thread string `json:",omitempty"`

	Name      json.Number `json:"tim,Number,omitempty"`
	Extension string      `json:"ext,omitempty"`
}

func (c4p Chan4Post) Board() string {
	return c4p.board
}

func (c4p Chan4Post) Thread() string {
	return c4p.thread
}

func (c4p Chan4Post) Files(excludeExtras bool) (files []File) {
	if c4p.Name != "" {
		files = append(files, Chan4File{c4p.board, c4p.thread, c4p.Name.String(), c4p.Extension[1:]})
	}
	return
}

type Chan4File struct {
	board     string
	thread    string
	name      string
	extension string
}

func (c4f Chan4File) URL() string {
	return fmt.Sprintf("https://i.4cdn.org/%s/%s.%s", c4f.board, c4f.name, c4f.extension)
}

func (c4f Chan4File) Board() string {
	return c4f.board
}

func (c4f Chan4File) Thread() string {
	return c4f.thread
}

func (c4f Chan4File) Name() string {
	return c4f.name
}

func (c4f Chan4File) Extension() string {
	return c4f.extension
}
