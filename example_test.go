package gorequests_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cairc/gorequests"
)

func Example_new() {
	text, err := gorequests.New(http.MethodGet, "https://httpbin.org/get").WithTimeout(time.Second * 10).Text()
	if err != nil {
		panic(err)
	}
	fmt.Println("text", text)
}

func Example_factory() {
	// I hope to set fixed parameters every time I initiate a request

	// Then, every request created by this factory will not log
	fac := gorequests.NewFactory(
		gorequests.WithLogger(gorequests.NewDiscardLogger()),
		gorequests.WithTimeout(time.Second*10),
	)

	// Send sample request
	text, err := fac.New(http.MethodGet, "https://httpbin.org/get").Text()
	if err != nil {
		panic(err)
	}
	fmt.Println("text", text)
}

func Example_newSession() {
	session := gorequests.NewSession("/tmp/gorequests-session.txt")
	text, err := session.New(http.MethodGet, "https://jsonplaceholder.typicode.com/todos/1").WithTimeout(time.Second * 10).Text()
	if err != nil {
		panic(err)
	}
	fmt.Println("text", text)
}
