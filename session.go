package gorequests

import (
	"net/http"
	"sync"

	cookiejar "github.com/juju/persistent-cookiejar"
)

type Session struct {
	jar        *cookiejar.Jar
	err        error
	cookiefile string
	options    []RequestOption
}

func (r *Session) New(method, url string) *Request {
	req := New(method, url)
	req.persistentJar = r.jar
	req.SetError(r.err)
	for _, v := range r.options {
		if err := v(req); err != nil {
			return req.SetError(err)
		}
	}
	return req
}

func (r *Session) Jar() http.CookieJar {
	return r.jar
}

func (r *Session) CookieFile() string {
	return r.cookiefile
}

var (
	sessionLock sync.Mutex
	sessionMap  map[string]*Session
)

func init() {
	sessionMap = map[string]*Session{}
}

// same cookie-file has same session instance
func NewSession(cookiefile string, options ...RequestOption) *Session {
	sessionLock.Lock()
	defer sessionLock.Unlock()

	v := sessionMap[cookiefile]
	if v != nil {
		return v
	}

	v = newSession(cookiefile, options)
	sessionMap[cookiefile] = v
	return v
}

func newSession(cookiefile string, options []RequestOption) *Session {
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: cookiefile,
	})
	if err != nil {
		return &Session{err: err, cookiefile: cookiefile, options: options}
	} else {
		return &Session{jar: jar, cookiefile: cookiefile, options: options}
	}
}
