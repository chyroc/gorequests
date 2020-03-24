package gorequests

import (
	"net/http"

	"github.com/juju/persistent-cookiejar"
)

type Session struct {
	*Request
}

func (r *Session) Get(url string) *Request {
	return New(http.MethodGet, url)
}

func (r *Session) Post(url string) *Request {
	return New(http.MethodPost, url)
}

func (r *Session) Put(url string) *Request {
	return New(http.MethodPut, url)
}

func (r *Session) Delete(url string) *Request {
	return New(http.MethodDelete, url)
}

func (r *Session) Connect(url string) *Request {
	return New(http.MethodConnect, url)
}

func (r *Session) Head(url string) *Request {
	return New(http.MethodHead, url)
}

func (r *Session) Patch(url string) *Request {
	return New(http.MethodPatch, url)
}

func (r *Session) Trace(url string) *Request {
	return New(http.MethodTrace, url)
}

func (r *Session) Options(url string) *Request {
	return New(http.MethodOptions, url)
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
