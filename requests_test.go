package gorequests_test

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/chyroc/gorequests"
	"github.com/stretchr/testify/assert"
)

func joinHttpBinURL(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return "https://httpbin.org" + path
}

func Test_Real(t *testing.T) {
	as := assert.New(t)

	t.Run("/ip", func(t *testing.T) {
		resp := struct {
			Origin string `json:"origin"`
		}{}
		err := gorequests.New(http.MethodGet, joinHttpBinURL("/ip")).Unmarshal(&resp)
		as.Nil(err)
		as.NotEmpty(resp.Origin)
	})

	t.Run("/user-agent", func(t *testing.T) {
		resp := struct {
			UserAgent string `json:"user-agent"`
		}{}
		err := gorequests.New(http.MethodGet, joinHttpBinURL("/user-agent")).Unmarshal(&resp)
		as.Nil(err)
		as.True(regexp.MustCompile(`gorequests/v\d+.\d+.\d+ \(https://github.com/chyroc/gorequests\)`).MatchString(resp.UserAgent),
			fmt.Sprintf("%s not match user-agent", resp.UserAgent))
	})

	t.Run("/headers", func(t *testing.T) {
		resp := struct {
			A string `json:"A"`
			B string `json:"B"`
		}{}
		as.Nil(gorequests.New(http.MethodGet, joinHttpBinURL("/headers")).WithHeader(
			"a", "1",
		).WithHeaders(map[string]string{
			"a": "2",
			"b": "3",
		}).Unmarshal(&resp))
		as.Equal("1,2", resp.A)
		as.Equal("3", resp.B)
	})

	t.Run("/get", func(t *testing.T) {
		resp := struct {
			Args struct {
				A []string `json:"a"`
				B string   `json:"b"`
			} `json:"args"`
		}{}
		as.Nil(gorequests.New(http.MethodGet, joinHttpBinURL("/get")).
			WithQuery("a", "1").WithQuerys(map[string]string{
			"a": "2",
			"b": "3",
		}).Unmarshal(&resp))
		as.Equal([]string{"1", "2"}, resp.Args.A)
		as.Equal("3", resp.Args.B)
	})

	t.Run("/status", func(t *testing.T) {
		status, err := gorequests.New(http.MethodGet, joinHttpBinURL("/status/403")).ResponseStatus()
		as.Nil(err)
		as.Equal(403, status)
	})

	t.Run("/delay/3", func(t *testing.T) {
		text, err := gorequests.New(http.MethodGet, joinHttpBinURL("/delay/4")).WithTimeout(time.Second).Text()
		as.Empty(text)
		as.NotNil(err)
		as.Contains(err.Error(), "context deadline exceeded")
	})

	t.Run("/image", func(t *testing.T) {
		t.Skip()

		gorequests.New(http.MethodGet, joinHttpBinURL("/image")).Text()
	})
}
