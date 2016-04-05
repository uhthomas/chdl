package board

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func DetailChan7(u *url.URL) (board, thread string, err error) {
	if u.Path == "/read.php" {
		q := u.Query()
		board = q.Get("b")
		thread = q.Get("t")
		if board == "" {
			err = ErrInvalidURLFormat
		}
		return
	}

	s := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(s) > 0 {
		board = s[0]
	}
	if len(s) > 2 {
		thread = strings.Split(s[2], ".")[0]
	}
	return
}

type Chan7 struct {
	board string
}

func NewChan7(u *url.URL) (Chan7, error) {
	board, _, err := DetailChan7(u)
	return Chan7{board}, err
}

func (c7 Chan7) Board() string {
	return c7.board
}

func (c7 Chan7) Thread(thread string) Thread {
	return Chan7Thread{c7.board, thread}
}

func (c7 Chan7) Threads() (threads []Thread, err error) {
	for i := 0; i < 8; i++ {
		page, err := c7.Page(i + 1)
		if err != nil {
			return nil, err
		}

		threads = append(threads, page...)
	}
	return
}

func (c7 Chan7) Page(i int) (threads []Thread, err error) {
	u := fmt.Sprintf("https://7chan.org/%s", c7.board)
	if i > 1 {
		u = fmt.Sprintf("%s/%d.json", u, i)
	}
	doc, err := goquery.NewDocument(u)
	if err != nil {
		return nil, err
	}

	doc.Find(".op > .post").Each(func(i int, s *goquery.Selection) {
		id, ok := s.Attr("id")
		if !ok {
			return
		}

		threads = append(threads, Chan7Thread{
			board:  c7.board,
			thread: id,
		})
	})
	return
}

func (c7 Chan7) Posts() (posts []Post, err error) {
	threads, err := c7.Threads()
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

func (c7 Chan7) Files(excludeExtras bool) (files []File, err error) {
	threads, err := c7.Threads()
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

type Chan7Thread struct {
	board  string
	thread string
}

func (c7t Chan7Thread) URL() string {
	return fmt.Sprintf("https://7chan.org/read.php?%s", url.Values{
		"b": {c7t.board},
		"t": {c7t.thread},
		"p": {"p1--"},
	}.Encode())
}

func (c7t Chan7Thread) Board() string {
	return c7t.board
}

func (c7t Chan7Thread) Thread() string {
	return c7t.thread
}

func (c7t Chan7Thread) Posts() (posts []Post, err error) {
	doc, err := goquery.NewDocument(c7t.URL())
	if err != nil {
		return nil, err
	}
	doc.Find(".post").Each(func(i int, s *goquery.Selection) {
		post := Chan7Post{board: c7t.board, thread: c7t.thread}

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

func (c7t Chan7Thread) Files(excludeExtras bool) (files []File, err error) {
	posts, err := c7t.Posts()
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		files = append(files, post.Files(excludeExtras)...)
	}
	return
}

type Chan7Post struct {
	board  string
	thread string

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

func (c7p Chan7Post) Thread() string {
	return c7p.thread
}

func (c7p Chan7Post) Files(excludeExtras bool) (files []File) {
	if c7p.Name != "" {
		files = append(files, Chan7File{c7p.board, c7p.thread, c7p.Name, c7p.Extension})
	}
	if c7p.Extras == nil || excludeExtras {
		return
	}

	for _, extra := range c7p.Extras {
		files = append(files, Chan7File{c7p.board, c7p.thread, extra.Name, extra.Extension})
	}
	return
}

type Chan7File struct {
	board     string
	thread    string
	name      string
	extension string
}

func (c7f Chan7File) URL() string {
	return fmt.Sprintf("https://7chan.org/%s/src/%s.%s", c7f.board, c7f.name, c7f.extension)
}

func (c7f Chan7File) Board() string {
	return c7f.board
}

func (c7f Chan7File) Thread() string {
	return c7f.thread
}

func (c7f Chan7File) Name() string {
	return c7f.name
}

func (c7f Chan7File) Extension() string {
	return c7f.extension
}
