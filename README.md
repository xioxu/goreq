# goreq
[![Build Status](https://travis-ci.org/xioxu/goreq.svg?branch=master)](https://travis-ci.org/xioxu/goreq)
[![GoDoc](https://godoc.org/github.com/xioxu/goreq?status.svg)](https://godoc.org/github.com/xioxu/goreq)
[![LICENSE](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://github.com/huaweicloud/golangsdk/blob/master/LICENSE)

A simple http request library for golang. 

# Install
`go get github.com/xioxu/goreq`

# Simple to use

```go
    req := goreq.Req(nil)
	body,_,_ := req.Get("https://www.baidu.com").Do()
	fmt.Print(string(body))
```

## Table of contents
- [Options](#options)
- [Proxy](#proxy)
- [Pipe](#pipe)
- [BodyContent](#bodycontent)

## Options

| Property        | Description           | Notes  |
| ------------- | ------------- | ----- |
| Method      | HTTP method | - |
| URL      | Fully qualified uri | - |
| Proxy      | Address of an HTTP proxy | - |
| Headers      | HTTP headers to setup for the request | - |
| FollowRedirect      | Follow HTTP 3xx responses as redirects | default: true , you can use goreq.TrueVal or goreq.FalseVal to set |
| Jar      | A cookie Jar to use for multiple subsequent requests | You can define your own custom cookie jar; see cookies section in following |
| QueryString      | Object containing querystring values to be appended to the uri | - |
| Timeout      | Request timeout value | Global request timeout(e.g: 15 * time.Second),  see https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/#clienttimeouts |
| HeadersToBeRemove      | The headers you want to remove before send request | - |

## Proxy
If you specify a proxy option, then the request (and any subsequent redirects) will be sent via a connection to the proxy server.

```go
    req := goreq.Req(&goreq.ReqOptions{ Proxy: goreq.NewString("http://localhost:8888")})
	body,resp,_ := req.Get("https://www.baidu.com").Do()
	fmt.Println(string(body))
	fmt.Println(resp.StatusCode)
```

## Pipe
There are some Pipe methods to handle different case:

### PipeStream
You can pipe any response to a writer. (Refer to UT: TestPipeSream)
```go
  req := goreq.Req(nil)
  req.Get("https://www.baidu.com").PipeStream(fileStream)
```

### PipeReq
You can pipe a request result to next request (Refer to UT: TestPipeReq)
```go
  req1 := goreq.Req(nil)
  req2 := goreq.Req(nil)
  req2.Post("http://www.bbb.com/submit")
  req.Get("https://www.baidu.com").PipeReq(req2)
```

### PipeFromReq
Pipe the http.Request content to goReq request (UT: TestPipeFromReq)

### PipeToResponse
Pipe the result to a http.Response

We can create a reverseProxy server through PipeFromReq and PipeToResponse easily:
```go
if err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		req := goreq.Req(&goreq.ReqOptions{
			Method: r.Method,
			Url:    "https://www.baidu.com" + r.RequestURI,
		})

		req.PipeFromReq(r).PipeToResponse(w)
	})); err != nil {
		panic(err)
	}
```

## BodyContent
You can set the request body with the folloing methods:

- JsonObject
- JsonString
- FormData 

## Cookie

```go
cookieJar,_ := cookiejar.New(nil)
req := goreq.Req(&goreq.ReqOptions{
			Method: "get",
			Url:    "https://www.baidu.com",
			Jar:cookieJar,
		})
```
