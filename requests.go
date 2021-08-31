package gorequests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	cookiejar "github.com/juju/persistent-cookiejar"
)

type Request struct {
	// internal
	cachedurl     string
	persistentJar *cookiejar.Jar
	lock          sync.RWMutex
	err           error
	logger        Logger

	// request
	context      context.Context     // request context
	isIgnoreSSL  bool                // request  ignore ssl verify
	header       http.Header         // request header
	querys       map[string][]string // request query
	isNoRedirect bool                // request ignore redirect
	timeout      time.Duration       // request timeout
	url          string              // request url
	method       string              // request method
	body         io.Reader           // request body

	// resp
	resp      *http.Response
	bytes     []byte
	isRead    bool
	isRequest bool
}

func New(method, url string) *Request {
	r := &Request{
		url:     url,
		method:  method,
		header:  map[string][]string{},
		querys:  make(map[string][]string),
		context: context.TODO(),
		logger: newDefaultLogger(),
	}
	r.header.Set("user-agent", fmt.Sprintf("gorequests/%s (https://github.com/chyroc/gorequests)", version))
	return r
}

func (r *Request) SetError(err error) *Request {
	r.err = err
	return r
}
