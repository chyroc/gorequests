package gorequests

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Request struct {
	Timeout time.Duration
	URL     string
	Method  string
	Body    io.Reader

	// req
	headers map[string]string

	// resp
	resp      *http.Response
	bytes     []byte
	isRead    bool
	isRequest bool

	// control
	lock sync.Mutex
	err  error
	jar  *cookiejar.Jar
}

func New(method, url string) *Request {
	return &Request{
		URL:     url,
		Method:  method,
		headers: make(map[string]string),
	}
}

// header
func (r *Request) WithHeader(k, v string) *Request {
	r.headers[k] = v
	return r
}

// header
func (r *Request) WithHeaders(kv map[string]string) *Request {
	for k, v := range kv {
		r.headers[k] = v
	}
	return r
}

// header
func (r *Request) ReqHeaders() map[string]string {
	return r.headers
}

// header
func (r *Request) RespHeaders() (map[string]string, error) {
	if r.err != nil {
		return nil, r.err
	}

	if err := r.doRequest(); err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for k, v := range r.resp.Header {
		m[k] = v[0]
	}
	return m, nil
}

// body
func (r *Request) WithBody(body interface{}) *Request {
	switch v := body.(type) {
	case io.Reader:
		r.Body = v
	case []byte:
		r.Body = bytes.NewReader(v)
	case string:
		r.Body = strings.NewReader(v)
	default:
		bs, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		r.Body = bytes.NewReader(bs)
	}

	return r
}

func (r *Request) Text() (string, error) {
	if r.err != nil {
		return "", r.err
	}

	if err := r.doRead(); err != nil {
		return "", err
	}

	return string(r.bytes), nil
}

func (r *Request) Bytes() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}

	if err := r.doRead(); err != nil {
		return nil, err
	}

	return r.bytes, nil
}

func (r *Request) Unmarshal(val interface{}) error {
	if r.err != nil {
		return r.err
	}

	bs, err := r.Bytes()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, val); err != nil {
		return errors.Errorf("unmarshal %s to %s failed: %s", bs, reflect.TypeOf(val).Name(), err)
	}
	return nil
}

func (r *Request) doRead() error {
	if r.err != nil {
		return r.err
	}

	if err := r.doRequest(); err != nil {
		return err
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	isRead := r.isRead
	if isRead {
		return nil
	}

	var err error
	r.bytes, err = ioutil.ReadAll(r.resp.Body)
	if err != nil {
		return errors.Errorf("read request(%s: %s) response failed: %w", r.Method, r.URL, err)
	}
	logrus.Infof("[gorequests] %s: %s, doRead: %s", r.Method, r.URL, r.bytes)
	r.isRead = true

	return nil
}

func (r *Request) doRequest() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.err != nil {
		return r.err
	}

	isRequest := r.isRequest
	if isRequest {
		return nil
	}

	logrus.Infof("[gorequests] %s: %s", r.Method, r.URL)

	if r.jar != nil {
		defer func() {
			if err := r.jar.Save(); err != nil {
				r.err = err
			}
		}()
	}

	req, err := http.NewRequest(r.Method, r.URL, r.Body)
	if err != nil {
		return errors.Errorf("new request(%s: %s) failed: %w", r.Method, r.URL, err)
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	c := http.Client{
		Jar:     r.jar,
		Timeout: r.Timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		return errors.Errorf("do request(%s: %s) failed: %w", r.Method, r.URL, err)
	}
	r.resp = resp
	r.isRequest = true
	return nil
}
