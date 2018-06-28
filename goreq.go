package goreq

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

var defaultTransport http.RoundTripper = &http.Transport{MaxIdleConns: 10, IdleConnTimeout: 30 * time.Second}
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
	Headers map[string][]string

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

type reqPipeContent struct {
	reader      io.ReadCloser
	contentType string
}

func (reqPipe *reqPipeContent) build() (contentType string, data io.Reader) {
	return reqPipe.contentType, reqPipe.reader
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
			url += "&" + qs
		} else {
			url += "?" + qs
		}
	}

	return url
}

func DefaultOptions() *ReqOptions {
	options := ReqOptions{
		Method:         "Get",
		FollowRedirect: false,
		Headers:        make(map[string][]string),
		QueryString:    make(url.Values),
	}

	return &options
}

func Options(opts *ReqOptions) *ReqOptions {
	return mergeOptions(opts, defaultOptions)
}

func mergeOptions(copyTo *ReqOptions, copyFrom *ReqOptions) *ReqOptions {
	if copyTo == nil {
		tmpOptions := *copyFrom
		return &tmpOptions
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
	} else {
		for k, v := range copyFrom.Headers {
			copyTo.Headers[k] = v
		}
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
	req.Options.Url = url
	req.Options.Method = "POST"
	return req
}

func (req *GoReq) Get(url string) *GoReq {
	req.Options.Url = url
	req.Options.Method = "Get"
	return req
}

func (req *GoReq) FormData(formData url.Values) *GoReq {
	req.Options.bodyContent = &formContent{
		content: formData,
	}
	return req
}

func (req *GoReq) JsonString(jsonStr []byte) *GoReq {
	req.Options.bodyContent = &jsonContent{
		content: jsonStr,
	}

	return req
}

func (req *GoReq) JsonObject(jsonObj interface{}) *GoReq {
	req.Options.bodyContent = &jsonObjContent{
		content: jsonObj,
	}

	return req
}

func (req *GoReq) PipeFromReq(r *http.Request) *GoReq {
	removeReqHeaders := map[string]interface{}{
		"Connection": 1,
	}
	pHeaders := make(map[string][]string)
	for k, v := range r.Header {

		if removeReqHeaders[k] == nil {
			pHeaders[k] = v
		}
	}

	req.Options = mergeOptions(req.Options, &ReqOptions{Headers: pHeaders})
	req.Options.bodyContent = &reqPipeContent{reader: r.Body, contentType: r.Header.Get("Content-Type")}

	return req
}

func (req *GoReq) PipeStream(writer io.Writer) error {
	reader, _, err := req.prepareReq()

	if err != nil {
		return err
	}
	defer (reader).Close()

	p := make([]byte, 4)

	for {
		n, err := (reader).Read(p)

		if err != nil {
			if err == io.EOF {
				(writer).Write(p[:n])
				break
			}

			return err
		}

		(writer).Write(p[:n])
	}

	return nil
}

func (req *GoReq) PipeReq(nextReq *GoReq) (*GoReq, error) {
	reader, resp, err := req.prepareReq()

	if err != nil {
		return nil, err
	}

	nextReq.Options.bodyContent = &reqPipeContent{reader: reader, contentType: resp.Header.Get("reqPipeContent")}
	return nextReq, nil
}

func (req *GoReq) To(result interface{}) (*http.Response, error) {
	body, resp, err := req.Do()

	if err == nil {
		err = json.Unmarshal((body), result)
	}

	return resp, err
}

func (req *GoReq) prepareReq() (io.ReadCloser, *http.Response, error) {
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

		if closer, ok := submitBody.(io.ReadCloser); ok {
			defer closer.Close()
		}
		req.Options.Headers["Content-Type"] = []string{contentType}
	}

	httpReq, err := http.NewRequest(strings.ToUpper(req.Options.Method), req.Options.buidUrl(), submitBody)
	if err != nil {
		return nil, nil, err
	}

	if req.Options.Headers != nil {
		for k, v := range req.Options.Headers {
			httpReq.Header[k] = v
		}
	}

	if err != nil {
		return nil, nil, err
	}

	resp, err := req.client.Do(httpReq)

	if err != nil {
		return nil, nil, err
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	return reader, resp, nil
}

func (req *GoReq) Do() ([]byte, *http.Response, error) {
	reader, resp, err := req.prepareReq()

	if err != nil {
		return nil, nil, err
	}
	defer (reader).Close()

	body, err := ioutil.ReadAll(reader)

	if err != nil {
		return nil, nil, err
	}

	return body, resp, nil
}
