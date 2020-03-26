package gorequests

import (
	cookiejar "github.com/juju/persistent-cookiejar"
)

type Session struct {
	jar *cookiejar.Jar
}

func (r *Session) New(method, url string) *Request {
	req := New(method, url)
	req.persistentJar = r.jar
	return req
}

func NewSession(cookiefile string) (*Session, error) {
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: cookiefile,
	})
	if err != nil {
		return nil, err
	}

	return &Session{
		jar: jar,
	}, nil
}
