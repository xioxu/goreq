[![Build Status](https://travis-ci.org/xioxu/goreq.svg?branch=master)](https://travis-ci.org/xioxu/goreq)
[![GoDoc](https://godoc.org/github.com/xioxu/goreq?status.svg)](https://godoc.org/github.com/xioxu/goreq)

# goreq
一个golang下的轻量级的http request工具

A simple http request library for golang.

# Install
`go get github.com/xioxu/goreq`

# Simple to use

``` go
    req := goreq.Req(nil)
	body,_,_ := req.Get("https://www.baidu.com").Do()
	fmt.Print(string(body))
```

## Table of contents
- [Proxies](#proxies)

## Proxies

如果你给option设置了Proxy值，那么http请求将会通过该代理进行转发

If you specify a proxy option, then the request (and any subsequent redirects) will be sent via a connection to the proxy server.

```go
    req := goreq.Req(&goreq.ReqOptions{Proxy:"http://localhost:8888"})
	body,resp,_ := req.Get("https://www.baidu.com").Do()
	fmt.Println(string(body))
	fmt.Println(resp.StatusCode)
```

## All avaliable options
- Proxy - A proxy address used for send http request
- Headers - The headers you want to add to request
- HeadersToBeRemove - The headers you want to remove before send request

