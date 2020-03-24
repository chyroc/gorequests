package gorequests

import (
	"github.com/juju/persistent-cookiejar"
)

type Session struct {
	*Request
}

func (r *Session) New(method, url string) *Request {
	return New(method, url)
}

func NewSession(cookiefile string) *Session {
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: cookiefile,
	})

	return &Session{
		Request: &Request{
			err: err,
			jar: jar,
		},
	}
}
