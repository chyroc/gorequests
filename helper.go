package gorequests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

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

	return writer.FormDataContentType(), body, nil
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

// request url
func (r *Request) parseRequestURL() string {
	URL, err := url.Parse(r.url)
	if err != nil {
		return r.url
	}
	q := URL.Query()
	for k, v := range r.querys {
		q[k] = append(q[k], v...)
	}
	URL.RawQuery = q.Encode()
	return URL.String()
}

func toBody(body interface{}) ([]byte, io.Reader, error) {
	switch v := body.(type) {
	case io.Reader:
		return nil, v, nil
	case []byte:
		return v, bytes.NewReader(v), nil
	case string:
		return []byte(v), strings.NewReader(v), nil
	default:
		bs, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		return bs, bytes.NewReader(bs), nil
	}
}

type s struct {
	idx   int
	query string
}

var queryToMapKeys sync.Map
