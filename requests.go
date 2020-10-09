package gorequests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
	url     string // use Request.URL() to access url
	Method  string
	Body    io.Reader

	// req
	headers       map[string]string
	querys        map[string][]string
	cachedurl     string
	isNoRedirect  bool
	persistentJar *cookiejar.Jar

	// resp
	resp      *http.Response
	bytes     []byte
	isRead    bool
	isRequest bool

	// control
	reqlock  sync.RWMutex
	readlock sync.Mutex
	err      error
}

func New(method, url string) *Request {
	return &Request{
		url:     url,
		Method:  method,
		headers: make(map[string]string),
		querys:  make(map[string][]string),
	}
}

// header
func (r *Request) WithHeader(k, v string) *Request {
	r.headers[k] = v
	return r
}

// 重定向，默认是 true
func (r *Request) WithRedirect(b bool) *Request {
	r.isNoRedirect = !b
	return r
}

// header
func (r *Request) WithHeaders(kv map[string]string) *Request {
	for k, v := range kv {
		r.headers[k] = v
	}
	return r
}

// query
func (r *Request) WithQuery(k, v string) *Request {
	r.reqlock.Lock()
	defer r.reqlock.Unlock()
	if r.cachedurl != "" {
		r.err = fmt.Errorf("[gorequests] already send request, cannot add query param")
		return r
	}
	r.querys[k] = append(r.querys[k], v)
	return r
}

// querys
func (r *Request) WithQuerys(kv map[string]string) *Request {
	r.reqlock.Lock()
	defer r.reqlock.Unlock()
	if r.cachedurl != "" {
		r.err = fmt.Errorf("[gorequests] already send request, cannot add query param")
		return r
	}
	for k, v := range kv {
		r.querys[k] = append(r.querys[k], v)
	}
	return r
}

// header
func (r *Request) Headers() map[string]string {
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
			r.err = err
			return r
		}
		r.Body = bytes.NewReader(bs)
	}

	return r
}

// request url
func (r *Request) RequestURL() string {
	r.reqlock.RLock()
	defer r.reqlock.RUnlock()

	r.parseURLInLock()
	return r.cachedurl
}

// request url
func (r *Request) parseURLInLock() {
	if r.cachedurl != "" {
		return
	}
	URL, err := url.Parse(r.url)
	if err != nil {
		r.cachedurl = r.url
		return
	}
	q := URL.Query()
	for k, v := range r.querys {
		q[k] = append(q[k], v...)
	}
	URL.RawQuery = q.Encode()
	r.cachedurl = URL.String()
	return
}

func (r *Request) SetError(err error) *Request {
	r.err = err
	return r
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

func (r *Request) Map() (map[string]interface{}, error) {
	if r.err != nil {
		return nil, r.err
	}

	bs, err := r.Bytes()
	if err != nil {
		return nil, err
	}

	var m = make(map[string]interface{})
	if err := json.Unmarshal(bs, &m); err != nil {
		return nil, errors.Wrapf(err, "unmarshal resp(%s) failed", bs)
	}
	return m, nil
}

func (r *Request) Text() (string, error) {
	if r.err != nil {
		return "", r.err
	}

	bs, err := r.Bytes()
	if err != nil {
		return "", err
	}

	return string(bs), nil
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

func (r *Request) doRead() error {
	if r.err != nil {
		return r.err
	}

	if err := r.doRequest(); err != nil {
		return err
	}

	r.readlock.Lock()
	defer r.readlock.Unlock()

	isRead := r.isRead
	if isRead {
		return nil
	}

	var err error
	r.bytes, err = ioutil.ReadAll(r.resp.Body)
	if err != nil {
		return errors.Wrapf(err, "read request(%s: %s) response failed", r.Method, r.cachedurl)
	}
	logrus.Debugf("[gorequests] %s: %s, doRead: %s", r.Method, r.cachedurl, r.bytes)
	r.isRead = true

	return nil
}

func (r *Request) doRequest() error {
	if r.err != nil {
		return r.err
	}

	r.reqlock.Lock()
	defer r.reqlock.Unlock()

	isRequest := r.isRequest
	if isRequest {
		return nil
	}

	r.parseURLInLock() // .url -> .cacheurl

	logrus.Debugf("[gorequests] %s: %s", r.Method, r.cachedurl)

	if r.persistentJar != nil {
		defer func() {
			if err := r.persistentJar.Save(); err != nil {
				_ = err // TODO: logs
			}
		}()
	}

	req, err := http.NewRequest(r.Method, r.cachedurl, r.Body)
	if err != nil {
		return errors.Wrapf(err, "new request(%s: %s) failed", r.Method, r.cachedurl)
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	c := &http.Client{
		Timeout: r.Timeout,
	}
	if r.persistentJar != nil {
		c.Jar = r.persistentJar
	}
	if r.isNoRedirect {
		c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	resp, err := c.Do(req)
	if err != nil {
		return errors.Wrapf(err, "do request(%s: %s) failed", r.Method, r.cachedurl)
	}
	r.resp = resp
	r.isRequest = true
	return nil
}

func (r *Request) Response() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}

	if err := r.doRequest(); err != nil {
		return nil, err
	}

	return r.resp, nil
}

func (r *Request) ResponseStatus() (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	if err := r.doRequest(); err != nil {
		return 0, err
	}

	return r.resp.StatusCode, nil
}
