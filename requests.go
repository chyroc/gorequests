package gorequests

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	cookiejar "github.com/juju/persistent-cookiejar"
)

type Request struct {
	// internal params
	lock         sync.RWMutex
	err          error
	startRequest bool

	// request params
	context      context.Context     // request context
	isIgnoreSSL  bool                // request  ignore ssl verify
	header       http.Header         // request header
	querys       map[string][]string // request query
	isNoRedirect bool                // request ignore redirect
	timeout      time.Duration       // request timeout

	url    string // use Request.URL() to access url
	Method string
	Body   io.Reader

	// internal
	cachedurl     string
	persistentJar *cookiejar.Jar

	// resp
	resp      *http.Response
	bytes     []byte
	isRead    bool
	isRequest bool
}

func New(method, url string) *Request {
	return &Request{
		url:    url,
		Method: method,
		header: map[string][]string{},
		querys: make(map[string][]string),
	}
}

func (r *Request) SetError(err error) *Request {
	r.err = err
	return r
}
