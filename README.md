# gorequests

[![codecov](https://codecov.io/gh/cairc/gorequests/branch/master/graph/badge.svg?token=Z73T6YFF80)](https://codecov.io/gh/cairc/gorequests)
[![go report card](https://goreportcard.com/badge/github.com/cairc/gorequests "go report card")](https://goreportcard.com/report/github.com/cairc/gorequests)
[![test status](https://github.com/cairc/gorequests/actions/workflows/test.yml/badge.svg)](https://github.com/cairc/gorequests/actions)
[![Apache-2.0 license](https://img.shields.io/badge/License-Apache%202.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/cairc/gorequests)
[![Go project version](https://badge.fury.io/go/github.com%2Fcairc%2Fgorequests.svg)](https://badge.fury.io/go/github.com%2Fcairc%2Fgorequests)

> Simple and easy-to-use go http client, supports cookies, streaming calls, custom logs and other functions

## Install

```shell
go get github.com/cairc/gorequests
```

## Usage

### Simple Send Request

```go
func main() {
	text, err := gorequests.New(http.MethodGet, "https://jsonplaceholder.typicode.com/todos/1").Text()
	if err != nil {
		panic(err)
	}
	fmt.Println("text", text)
}
```

### Send Request With Cookie

```go
func main() {
	session := gorequests.NewSession("/tmp/gorequests-session.txt")
	text, err := session.New(http.MethodGet, "https://jsonplaceholder.typicode.com/todos/1").Text()
	if err != nil {
		panic(err)
	}
	fmt.Println("text", text)
}
```

### Request Factory

```go
func main() {
    fac := gorequests.NewFactory(
        gorequests.WithLogger(gorequests.NewDiscardLogger()),
    )
	text, err := fac.New(http.MethodGet, "https://jsonplaceholder.typicode.com/todos/1").Text()
	if err != nil {
		panic(err)
	}
	fmt.Println("text", text)
}
```
