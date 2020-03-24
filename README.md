# gorequests

> 简单易用的带有 session 功能的 go http 客户端

## 支持功能

- 设置 header
- 设置 body
- 获取返回 text、bytes、map、interface


## 入手指南

### 简单使用

```go
func Example_Request() {
	text, err := gorequests.New(http.MethodGet, "https://jsonplaceholder.typicode.com/todos/1").Text()
	if err != nil {
		panic(err)
	}
	fmt.Println("text", text)
}
```

### 带 session 的请求

```go
func Example_Session() {
	session := gorequests.NewSession("/tmp/gorequests-session.txt")
	text, err := session.New(http.MethodGet, "https://jsonplaceholder.typicode.com/todos/1").Text()
	if err != nil {
		panic(err)
	}
	fmt.Println("text", text)
}
```
