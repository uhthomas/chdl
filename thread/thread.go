package thread

import (
	"errors"
	"net/url"
)

var (
	ErrUnknownChan      = errors.New("unknown or unsupported chan")
	ErrInvalidURLFormat = errors.New("invalid chan url format")
)

type Thread interface {
	URL() string
	Board() string
	Thread() string
	Posts() ([]Post, error)
	Files(bool) ([]File, error)
}

type Post interface {
	Board() string
	Files(bool) []File
}

type File interface {
	URL() string
	Board() string
	Name() string
	Extension() string
}

func New(u *url.URL) (Thread, error) {
	switch u.Host {
	case "boards.4chan.org":
		return NewChan4(u)
	case "7chan.org":
		return NewChan7(u)
	case "8ch.net":
		return NewChan8(u)
	}
	return nil, ErrUnknownChan
}
