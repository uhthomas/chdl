package board

import (
	"errors"
	"net/url"
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
	case "4chan.org", "boards.4chan.org":
		return NewChan4(u)
	case "7chan.org":
		return NewChan7(u)
	case "8ch.net":
		return NewChan8(u)
	}
	return nil, ErrUnknownChan
}

func Detail(u *url.URL) (board, thread string, err error) {
	switch u.Host {
	case "4chan.org", "boards.4chan.org":
		return DetailChan4(u)
	case "7chan.org":
		return DetailChan7(u)
	case "8ch.net":
		return DetailChan8(u)
	}
	return "", "", ErrUnknownChan
}
