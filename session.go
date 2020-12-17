package gorequests

import (
	cookiejar "github.com/juju/persistent-cookiejar"
)

type Session struct {
	jar *cookiejar.Jar
	err error
}

func (r *Session) New(method, url string) *Request {
	req := New(method, url)
	req.persistentJar = r.jar
	req.SetError(r.err)
	return req
}

func NewSession(cookiefile string) *Session {
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: cookiefile,
	})
	if err != nil {
		return &Session{err: err}
	} else {
		return &Session{jar: jar}
	}
}
