package gorequests_test

import (
	"fmt"
	"github.com/chyroc/gorequests"
	"net/http"
)

func Example_Request() {
	text, err := gorequests.New(http.MethodGet, "https://jsonplaceholder.typicode.com/todos/1").Text()
	if err != nil {
		panic(err)
	}
	fmt.Println("text", text)
}

func Example_Session() {
	session := gorequests.NewSession("/tmp/gorequests-session.txt")
	text, err := session.New(http.MethodGet, "https://jsonplaceholder.typicode.com/todos/1").Text()
	if err != nil {
		panic(err)
	}
	fmt.Println("text", text)
}
