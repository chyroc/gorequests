/*
Package gorequests send http request with human style.

for send GET request:
    gorequests.New(http.MethodGet, "https://httpbin.org/get")

for send POST request:
    gorequests.New(http.MethodPost, "https://httpbin.org/post).WithBody("text body")

for send http upload request
    gorequests.New(http.MethodPost, "https://httpbin.org/post).WithFile("1.txt", strings.NewReader("hi"), "file", nil)

for send json request
    gorequests.New(http.MethodPost, "https://httpbin.org/post).WithJSON(map[string]string{"key": "val"})

request with timeout
    gorequests.New(http.MethodGet, "https://httpbin.org/get).WithTimeout(time.Second)

request with context
    gorequests.New(http.MethodGet, "https://httpbin.org/get).WithContext(context.TODO())

request with no redirect
    gorequests.New(http.MethodGet, "https://httpbin.org/status/302).WithRedirect(false)

*/
package gorequests
