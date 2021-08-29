package gorequests

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/pkg/errors"
)

type Request struct {
	Context context.Context

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
	isIgnoreSSL   bool

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

// context
func (r *Request) WithCTX(ctx context.Context) *Request {
	if r.err != nil {
		return r
	}

	r.Context = ctx
	return r
}

// ignore ssl
func (r *Request) WithIgnoreSSL(ignore bool) *Request {
	if r.err != nil {
		return r
	}

	r.isIgnoreSSL = ignore
	return r
}

// header
func (r *Request) WithHeader(k, v string) *Request {
	if r.err != nil {
		return r
	}

	r.headers[k] = v
	return r
}

// 重定向，默认是 true
func (r *Request) WithRedirect(b bool) *Request {
	if r.err != nil {
		return r
	}

	r.isNoRedirect = !b
	return r
}

// header
func (r *Request) WithHeaders(kv map[string]string) *Request {
	if r.err != nil {
		return r
	}

	for k, v := range kv {
		r.headers[k] = v
	}
	return r
}

// query
func (r *Request) WithQuery(k, v string) *Request {
	if r.err != nil {
		return r
	}

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
	if r.err != nil {
		return r
	}

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

// query-struct
func (r *Request) WithQueryStruct(v interface{}) *Request {
	if r.err != nil {
		return r
	}

	r.reqlock.Lock()
	defer r.reqlock.Unlock()
	if r.cachedurl != "" {
		r.err = fmt.Errorf("[gorequests] already send request, cannot add query param")
		return r
	}
	kv, err := queryToMap(v)
	if err != nil {
		r.err = err
		return r
	}
	for k, v := range kv {
		r.querys[k] = append(r.querys[k], v...)
	}
	return r
}

// header
func (r *Request) Headers() map[string]string {
	return r.headers
}

// header
func (r *Request) GetHeaderString(key string) (string, error) {
	if r.err != nil {
		return "", r.err
	}

	if err := r.doRequest(); err != nil {
		return "", err
	}

	for k, v := range r.resp.Header {
		if key == k && len(v) > 0 {
			return v[0], nil
		}
	}
	return "", nil
}

func (r *Request) GetHeaderArray(key string) ([]string, error) {
	if r.err != nil {
		return nil, r.err
	}

	if err := r.doRequest(); err != nil {
		return nil, err
	}

	for k, v := range r.resp.Header {
		if key == k {
			return v, nil
		}
	}
	return nil, nil
}

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
	if r.err != nil {
		return r
	}

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

// body json
func (r *Request) WithJSON(body interface{}) *Request {
	if r.err != nil {
		return r
	}

	r.WithBody(body)
	r.headers["Content-Type"] = "application/json"

	return r
}

// body file
func (r *Request) WithFile(filename string, file io.Reader, fileKey string, params map[string]string) *Request {
	if r.err != nil {
		return r
	}

	contentType, bod, err := newFileUploadRequest(params, fileKey, filename, file)
	if err != nil {
		r.err = err
		return r
	}
	r.WithBody(bod)
	r.headers["Content-Type"] = contentType

	return r
}

// form data
func (r *Request) WithForm(body map[string]string) *Request {
	if r.err != nil {
		return r
	}

	buf := bytes.Buffer{}
	f := multipart.NewWriter(&buf)
	for k, v := range body {
		if err := f.WriteField(k, v); err != nil {
			r.err = err
			return r
		}
	}

	r.Body = strings.NewReader(buf.String())
	r.headers["Content-Type"] = f.FormDataContentType()

	return r
}

// cookie
func (r *Request) WithURLCookie(uri string) *Request {
	if r.err != nil {
		return r
	}
	if r.persistentJar == nil {
		return r
	}

	uriParse, err := url.Parse(uri)
	if err != nil {
		r.err = err
		return r
	}

	cookies := []string{}
	for _, v := range r.persistentJar.Cookies(uriParse) {
		cookies = append(cookies, v.Name+"="+v.Value)
	}
	if len(cookies) > 0 {
		r.headers["cookie"] = strings.Join(cookies, "; ")
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

	m := make(map[string]interface{})
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
	logger.Info(r.ctx(), "[gorequests] %s: %s, doRead: %s", r.Method, r.cachedurl, r.bytes)
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

	logger.Info(r.ctx(), "[gorequests] %s: %s", r.Method, r.cachedurl)

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
	if r.isIgnoreSSL {
		c.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
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

func (r *Request) ctx() context.Context {
	if r.Context != nil {
		return r.Context
	}
	return context.Background()
}

func queryToMap(v interface{}) (map[string][]string, error) {
	ss, err := getQueryToMapKeys(v)
	if err != nil {
		return nil, err
	} else if len(ss) == 0 {
		return map[string][]string{}, nil
	}

	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}

	vals := map[string][]string{}
	for _, s := range ss {
		vals[s.query], err = toStringList(vv.Field(s.idx))
		if err != nil {
			return nil, err
		}
	}

	return vals, nil
}

func toStringList(v reflect.Value) ([]string, error) {
	switch v.Kind() {
	case reflect.String:
		return []string{v.String()}, nil
	case reflect.Bool:
		if v.Bool() {
			return []string{"true"}, nil
		}
		return []string{"false"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []string{strconv.FormatInt(v.Int(), 10)}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []string{strconv.FormatUint(v.Uint(), 10)}, nil
	case reflect.Array, reflect.Slice:
		res := []string{}
		for j := 0; j < v.Len(); j++ {
			x, err := toStringList(v.Index(j))
			if err != nil {
				return nil, err
			}
			res = append(res, x...)
		}
		return res, nil
	}

	return nil, fmt.Errorf("invalid value: %s", v.Kind())
}

func getQueryToMapKeys(v interface{}) ([]s, error) {
	origin := reflect.TypeOf(v)
	v, ok := queryToMapKeys.Load(origin)
	if ok {
		return v.([]s), nil
	}

	vt := origin
	// vv := reflect.ValueOf(v)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
		// vv = vv.Elem()
	}
	if vt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("need strcut, but got %s", vt.Kind())
	}

	ss := []s{}
	for i := 0; i < vt.NumField(); i++ {
		itemT := vt.Field(i)
		// itemV := vv.Field(i)

		queryKey := itemT.Tag.Get("query")
		if queryKey == "" {
			continue
		}
		// itemV.String()
		ss = append(ss, s{
			idx:   i,
			query: queryKey,
		})
	}

	queryToMapKeys.Store(origin, ss)

	return ss, nil
}

type s struct {
	idx   int
	query string
}

var queryToMapKeys sync.Map

func newFileUploadRequest(params map[string]string, filekey, filename string, reader io.Reader) (string, io.Reader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(filekey, filename)
	if err != nil {
		return "", nil, err
	}
	if reader != nil {
		if _, err = io.Copy(part, reader); err != nil {
			return "", nil, err
		}
	}
	for key, val := range params {
		if err = writer.WriteField(key, val); err != nil {
			return "", nil, err
		}
	}
	if err = writer.Close(); err != nil {
		return "", nil, err
	}

	// fmt.Println("body",body.String())
	return writer.FormDataContentType(), body, nil
}
