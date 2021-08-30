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
	reqlock      sync.RWMutex
	err          error
	startRequest bool

	// request params
	context      context.Context     // request context
	isIgnoreSSL  bool                // request  ignore ssl verify
	header       http.Header         // request header
	querys       map[string][]string // request query
	isNoRedirect bool                // request ignore redirect

	Timeout time.Duration
	url     string // use Request.URL() to access url
	Method  string
	Body    io.Reader

	// req

	cachedurl     string
	persistentJar *cookiejar.Jar

	// resp
	resp      *http.Response
	bytes     []byte
	isRead    bool
	isRequest bool

	// control
	readlock sync.Mutex
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
