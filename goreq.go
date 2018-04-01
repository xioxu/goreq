package goreq

import (
	"time"
	"net/http"
	"io/ioutil"
	"net/url"
	"strings"
	"bytes"
	"io"
	"encoding/json"
	"net/http/cookiejar"
	"compress/gzip"
)

var defaultTransport http.RoundTripper = &http.Transport{MaxIdleConns: 10, IdleConnTimeout: 30 * time.Second,}
var defaultClient = &http.Client{Transport: defaultTransport}
var defaultOptions = DefaultOptions()

type GoReq struct {
	Options   *ReqOptions
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
	Jar *cookiejar.Jar

	//an HTTP proxy url to be used
	Proxy string

	//object containing querystring values to be appended to the uri
	QueryString url.Values

	bodyContent httpReqBody

	Timeout time.Duration
}

type httpReqBody interface {
	build() (contentType string, data io.Reader)
}

type formContent struct {
	content url.Values
}

func (form *formContent) build() (contentType string, data io.Reader) {
	return "application/x-www-form-urlencoded", strings.NewReader(form.content.Encode())
}

type jsonContent struct {
	content []byte
}

func (jsonString *jsonContent) build() (contentType string, data io.Reader) {
	return "application/json", bytes.NewReader(jsonString.content)
}

type jsonObjContent struct {
	content interface{}
}

func guard(err error) {
	if err != nil {
		panic(err)
	}
}

func (jsonObj *jsonObjContent) build() (contentType string, data io.Reader) {
	bytesContent, err := json.Marshal(jsonObj.content)
	guard(err)
	return "application/json", bytes.NewReader(bytesContent)
}

func (options *ReqOptions) buidUrl() string {
	url := options.Url

	qs := options.QueryString.Encode()

	if qs != "" {
		if strings.Contains(qs, "?") {
			url += "&" + qs;
		} else {
			url += "?" + qs;
		}
	}

	return url
}

func DefaultOptions() *ReqOptions {
	options := ReqOptions{
		Method:         "Get",
		FollowRedirect: false,
		Headers:        make(map[string]string),
		QueryString:    make(url.Values),
	}

	return &options
}

func Options(opts *ReqOptions) *ReqOptions {
	return mergeOptions(opts, defaultOptions)
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

	if copyTo.Jar == nil {
		copyTo.Jar = copyFrom.Jar
	}

	if copyTo.Headers == nil {
		copyTo.Headers = copyFrom.Headers
	}

	return copyTo
}

func Req(options *ReqOptions) *GoReq {
	goReq := GoReq{}
	goReq.transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	goReq.client = &http.Client{
		Transport: goReq.transport,
	}

	goReq.Options = mergeOptions(options, defaultOptions)
	return &goReq
}

func (req *GoReq) Post(url string) *GoReq {
	req.Options.Url = url;
	req.Options.Method = "POST"
	return req
}

func (req *GoReq) Get(url string) *GoReq {
	req.Options.Url = url;
	req.Options.Method = "Get"
	return req
}

func (req *GoReq) Form(formData url.Values) {
	req.Options.bodyContent = &formContent{
		content: formData,
	}
}

func (req *GoReq) JsonString(jsonStr []byte) {
	req.Options.bodyContent = &jsonContent{
		content: jsonStr,
	}
}

func (req *GoReq) JsonObject(jsonObj interface{}) {
	req.Options.bodyContent = &jsonObjContent{
		content: jsonObj,
	}
}

func (req *GoReq) Do() ([]byte, *http.Response, error) {
	if req.Options.Proxy != "" {
		parsedProxyUrl, err := url.Parse(req.Options.Proxy)

		if err != nil {
			return nil, nil, err
		} else {
			req.transport.Proxy = http.ProxyURL(parsedProxyUrl)
		}
	}

	if req.Options.Jar != nil {
		req.client.Jar = req.Options.Jar
	}

	if req.Options.Timeout > 0 {
		req.client.Timeout = req.Options.Timeout
	}

	var submitBody io.Reader
	var contentType string
	if req.Options.bodyContent != nil {
		contentType, submitBody = req.Options.bodyContent.build()
		req.Options.Headers["Content-Type"] = contentType
	}

	httpReq, err := http.NewRequest(strings.ToUpper(req.Options.Method), req.Options.buidUrl(), submitBody)
	if err != nil {
		return nil, nil, err
	}

	if req.Options.Headers != nil {
		for k, v := range req.Options.Headers {
			httpReq.Header.Add(k, v)
		}
	}

	resp, err := req.client.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, nil, err
	}

	return body, resp, nil
}
