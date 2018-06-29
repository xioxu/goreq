[![Build Status](https://travis-ci.org/xioxu/goreq.svg?branch=master)](https://travis-ci.org/xioxu/goreq)
[![GoDoc](https://godoc.org/github.com/xioxu/goreq?status.svg)](https://godoc.org/github.com/xioxu/goreq)

# goreq
A simple http request library for golang. 

golang下的一个轻量级http request类库

# Install
`go get github.com/xioxu/goreq`

# Simple to use

``` go
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
- Proxy - A proxy address used for send http request
- Headers - The headers you want to add to http request
- HeadersToBeRemove - The headers you want to remove before send request

## Proxy
If you specify a proxy option, then the request (and any subsequent redirects) will be sent via a connection to the proxy server.

```go
    req := goreq.Req(&goreq.ReqOptions{Proxy: &goreq.NullableString{Value:"http://localhost:8888"}})
	body,resp,_ := req.Get("https://www.baidu.com").Do()
	fmt.Println(string(body))
	fmt.Println(resp.StatusCode)
```

## Pipe
There are some Pipe methods to handle different case:

### PipeStream
You can pipe any response to a writer. (Refer to UT: TestPipeSream)
```
  req := goreq.Req(nil)
  req.Get("https://www.baidu.com").PipeStream(fileStream)
```

### PipeReq
You can pipe a request result to next request (Refer to UT: TestPipeReq)
```
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
```$xslt
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