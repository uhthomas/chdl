package thread

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

type Chan4 struct {
	url    *url.URL
	board  string
	thread string
}

func NewChan4(u *url.URL) (Chan4, error) {
	r, err := regexp.Compile(`^\/([a-z0-9]+?)\/thread\/([0-9]+?)(?:$|\/)`)
	if err != nil {
		return Chan4{}, err
	}

	matches := r.FindStringSubmatch(u.Path)
	if len(matches) != 3 {
		return Chan4{}, ErrInvalidURLFormat
	}
	return Chan4{u, matches[1], matches[2]}, nil
}

func (c4 Chan4) URL() string {
	return c4.url.String()
}

func (c4 Chan4) Board() string {
	return c4.board
}

func (c4 Chan4) Thread() string {
	return c4.thread
}

func (c4 Chan4) Posts() (posts []Post, err error) {
	res, err := http.Get(fmt.Sprintf("https://a.4cdn.org/%s/thread/%s.json", c4.board, c4.thread))
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
		post.board = c4.board
		posts = append(posts, post)
	}
	return
}

func (c4 Chan4) Files(excludeExtras bool) (files []File, err error) {
	posts, err := c4.Posts()
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		files = append(files, post.Files(excludeExtras)...)
	}
	return
}

type Chan4Post struct {
	board string `json:",omitempty"`

	Name      json.Number `json:"tim,Number,omitempty"`
	Extension string      `json:"ext,omitempty"`
	Extras    []struct {
		Name      string `json:"tim,Number"`
		Extension string `json:"ext"`
	} `json:"extra,omitempty"`
}

func (c4p Chan4Post) Board() string {
	return c4p.board
}

func (c4p Chan4Post) Files(excludeExtras bool) (files []File) {
	if c4p.Name != "" {
		files = append(files, Chan4File{c4p.board, c4p.Name.String(), c4p.Extension[1:]})
	}
	if c4p.Extras == nil || excludeExtras {
		return
	}

	for _, extra := range c4p.Extras {
		files = append(files, Chan4File{c4p.board, extra.Name, extra.Extension[1:]})
	}
	return
}

type Chan4File struct {
	board     string
	name      string
	extension string
}

func (c4f Chan4File) URL() string {
	return fmt.Sprintf("https://i.4cdn.org/%s/%s.%s", c4f.board, c4f.name, c4f.extension)
}

func (c4f Chan4File) Board() string {
	return c4f.board
}

func (c4f Chan4File) Name() string {
	return c4f.name
}

func (c4f Chan4File) Extension() string {
	return c4f.extension
}
