package goreq

import (
	"time"
	"net/http"
	"io/ioutil"
	"net/url"
)

var defaultTransport http.RoundTripper = &http.Transport{MaxIdleConns: 10, IdleConnTimeout: 30 * time.Second,}
var defaultClient = &http.Client{Transport: defaultTransport}
var defaultOptions = DefaultOptions()

type GoReq struct {
	options   *ReqOptions
	client    *http.Client
	transport *http.Transport
}

type ReqOptions struct {
	//http method (default: "GET")
	Method string

	//fully qualified uri or a parsed url object from url.parse()
	Url string

	//http headers (default: {})
	Headers map[string]string

	// follow HTTP 3xx responses as redirects (default: true).
	FollowRedirect bool

	// if not nil, remember cookies for future use (or define your custom cookie jar; see examples section)
	Jar *http.Cookie

	//an HTTP proxy url to be used
	Proxy string
}

func DefaultOptions() *ReqOptions {
	options := ReqOptions{
		Method:         "Get",
		FollowRedirect: false,
		Headers:make(map[string]string),
	}

	return &options
}

func Options(opts *ReqOptions)  *ReqOptions{
	return mergeOptions(opts,defaultOptions)
}

func request(url string, callback func(error, *http.Response, []byte)) {
	resp, err := defaultClient.Get(url)
	errHandler := func(error) bool {
		if err != nil {
			if callback != nil {
				callback(err, resp, nil)
			}
			return true
		}
		return false
	}

	if (!errHandler(err)) {
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		if (!errHandler(err)) {
			callback(err, resp, body)
		}
	}
}

func mergeOptions(copyTo *ReqOptions, copyFrom *ReqOptions) *ReqOptions {
	if copyTo == nil {
		return copyFrom
	}

	if copyFrom == nil {
		return copyTo
	}

	if copyTo.Method == "" {
		copyTo.Method = copyFrom.Method
	}

	if copyTo.Url == "" {
		copyTo.Url = copyFrom.Url
	}

	if copyTo.Jar != nil {
		copyTo.Jar = copyFrom.Jar
	}

	if copyTo.Headers == nil {
		copyTo.Headers = copyFrom.Headers
	}

	return copyTo
}

func Req(options *ReqOptions) *GoReq {
	goReq := GoReq{}
	goReq.transport = &http.Transport{}
	goReq.client = &http.Client{
		Transport: goReq.transport,
	}
	goReq.transport.Proxy = http.ProxyFromEnvironment
	goReq.options = mergeOptions(options, defaultOptions)
	return &goReq
}

func (req *GoReq) Post(url string) *GoReq {
	req.options.Url = url;
	req.options.Method = "POST"
	return  req
}

func (req *GoReq) Get(url string) *GoReq {
	req.options.Url = url;
	req.options.Method = "Get"
	return  req
}

func (req *GoReq) Do() (  []byte,*http.Response,error) {
	if req.options.Proxy != "" {
		parsedProxyUrl, err := url.Parse(req.options.Proxy)

		if err != nil {
			return nil, nil,err
		} else {
			req.transport.Proxy = http.ProxyURL(parsedProxyUrl)
		}
	}

	httpReq, err := http.NewRequest(req.options.Method, req.options.Url, nil)
	if err != nil {
		return nil, nil,err
	}

	if req.options.Headers != nil {
		for k, v := range req.options.Headers {
			httpReq.Header.Add(k, v)
		}
	}

	resp, err := req.client.Do(httpReq)
	if err != nil {
		return nil, nil,err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil,err
	}

	return  body,resp,nil
}
