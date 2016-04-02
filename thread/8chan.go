package thread

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

type Chan8 struct {
	url    *url.URL
	board  string
	thread string
}

func NewChan8(u *url.URL) (Chan8, error) {
	r, err := regexp.Compile(`^\/([a-z0-9]+?)\/res\/([0-9]+?)\.html`)
	if err != nil {
		return Chan8{}, err
	}

	matches := r.FindStringSubmatch(u.Path)
	if len(matches) != 3 {
		return Chan8{}, ErrInvalidURLFormat
	}

	return Chan8{u, matches[1], matches[2]}, nil
}

func (c8 Chan8) URL() string {
	return c8.url.String()
}

func (c8 Chan8) Board() string {
	return c8.board
}

func (c8 Chan8) Thread() string {
	return c8.thread
}

func (c8 Chan8) Posts() (posts []Post, err error) {
	res, err := http.Get(fmt.Sprintf("https://8ch.net/%s/res/%s.json", c8.board, c8.thread))
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
		post.board = c8.board
		posts = append(posts, post)
	}

	return
}

func (c8 Chan8) Files(excludeExtras bool) (files []File, err error) {
	posts, err := c8.Posts()
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		files = append(files, post.Files(excludeExtras)...)
	}

	return
}

type Chan8Post struct {
	board string `json:",omitempty"`

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
		files = append(files, Chan8File{c8p.board, c8p.Name.String(), c8p.Extension[1:]})
	}

	if c8p.Extras == nil || excludeExtras {
		return
	}

	for _, extra := range c8p.Extras {
		files = append(files, Chan8File{c8p.board, extra.Name, extra.Extension[1:]})
	}

	return
}

type Chan8File struct {
	board     string
	name      string
	extension string
}

func (c8f Chan8File) URL() string {
	return fmt.Sprintf("https://8ch.net/%s/src/%s.%s", c8f.board, c8f.name, c8f.extension)
}

func (c8f Chan8File) Board() string {
	return c8f.board
}

func (c8f Chan8File) Name() string {
	return c8f.name
}

func (c8f Chan8File) Extension() string {
	return c8f.extension
}
