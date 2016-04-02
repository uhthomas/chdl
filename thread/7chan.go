package thread

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Chan7 struct {
	url    *url.URL
	board  string
	thread string
}

func NewChan7(u *url.URL) (Chan7, error) {
	r, err := regexp.Compile(`^\/([a-z0-9]+?)\/res\/([0-9]+?)\.html`)
	if err != nil {
		q := u.Query()
		board := q.Get("b")
		thread := q.Get("t")
		if board == "" || thread == "" {
			return Chan7{}, ErrInvalidURLFormat
		}
		return Chan7{u, board, thread}, nil
	}

	matches := r.FindStringSubmatch(u.Path)
	if len(matches) != 3 {
		return Chan7{}, ErrInvalidURLFormat
	}

	return Chan7{u, matches[1], matches[2]}, nil
}

func (c7 Chan7) URL() string {
	return c7.url.String()
}

func (c7 Chan7) Board() string {
	return c7.board
}

func (c7 Chan7) Thread() string {
	return c7.thread
}

func (c7 Chan7) Posts() (posts []Post, err error) {
	u, err := url.Parse("https://7chan.org/read.php")
	if err != nil {
		return nil, err
	}

	u.Query().Set("b", c7.board)
	u.Query().Set("t", c7.thread)
	u.Query().Set("p", "p1--")

	doc, err := goquery.NewDocument(u.String())
	if err != nil {
		return nil, err
	}

	doc.Find(".post").Each(func(i int, s *goquery.Selection) {
		post := Chan7Post{board: c7.board}

		fs := s.Find(".file_size a")
		if fs.Length() == 1 {
			sp := strings.Split(strings.TrimSpace(fs.Text()), ".")
			post.Name = sp[0]
			post.Extension = sp[1]
		}

		s.Find("span.multithumbfirst a, span.multithumb a").Each(func(x int, ts *goquery.Selection) {
			iu, _ := ts.Attr("href")
			_, n := path.Split(iu)
			sp := strings.Split(strings.TrimSpace(n), ".")
			post.Extras = append(post.Extras, struct {
				Name      string
				Extension string
			}{sp[0], sp[1]})
		})

		posts = append(posts, post)
	})
	return
}

func (c7 Chan7) Files(excludeExtras bool) (files []File, err error) {
	posts, err := c7.Posts()
	if err != nil {
		return nil, err
	}
	for _, post := range posts {
		files = append(files, post.Files(excludeExtras)...)
	}
	return
}

type Chan7Post struct {
	board     string
	Name      string
	Extension string
	Extras    []struct {
		Name      string
		Extension string
	}
}

func (c7p Chan7Post) Board() string {
	return c7p.board
}

func (c7p Chan7Post) Files(excludeExtras bool) (files []File) {
	if c7p.Name != "" {
		files = append(files, Chan7File{c7p.board, c7p.Name, c7p.Extension})
	}
	if c7p.Extras == nil || excludeExtras {
		return
	}
	for _, extra := range c7p.Extras {
		files = append(files, Chan7File{c7p.board, extra.Name, extra.Extension})
	}
	return
}

type Chan7File struct {
	board     string
	name      string
	extension string
}

func (c7f Chan7File) URL() string {
	return fmt.Sprintf("https://7chan.org/%s/src/%s.%s", c7f.board, c7f.name, c7f.extension)
}

func (c7f Chan7File) Board() string {
	return c7f.board
}

func (c7f Chan7File) Name() string {
	return c7f.name
}

func (c7f Chan7File) Extension() string {
	return c7f.extension
}
