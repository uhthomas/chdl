package chdl

import (
	"errors"
	"net/url"

	"github.com/uhthomas/chdl/pkg"
)

var (
	ErrUnknownChan      = errors.New("unknown or unsupported chan")
	ErrInvalidURLFormat = errors.New("invalid chan url format")
)

type Board interface {
	Board() string
	Thread(string) Thread
	Threads() ([]Thread, error)
	Page(int) ([]Thread, error)
	Posts() ([]Post, error)
	Files(bool) ([]File, error)
}

type Thread interface {
	URL() string
	Board() string
	Thread() string
	Posts() ([]Post, error)
	Files(bool) ([]File, error)
}

type Post interface {
	Board() string
	Thread() string
	Files(bool) []File
}

type File interface {
	URL() string
	Board() string
	Thread() string
	Name() string
	Extension() string
}

func New(u *url.URL) (Board, error) {
	switch u.Host {
	case "boards.4chan.org", "www.4chan.org", "4chan.org":
		return pkg.NewChan4(u)
	case "www.7chan.org", "7chan.org":
		return pkg.NewChan7(u)
	case "www.8ch.net", "8ch.net":
		return pkg.NewChan8(u)
	}
	return nil, ErrUnknownChan
}

func Detail(u *url.URL) (board, thread string, err error) {
	switch u.Host {
	case "boards.4chan.org", "www.4chan.org", "4chan.org":
		return pkg.DetailChan4(u)
	case "www.7chan.org", "7chan.org":
		return pkg.DetailChan7(u)
	case "www.8ch.net", "8ch.net":
		return pkg.DetailChan8(u)
	}
	return "", "", ErrUnknownChan
}
