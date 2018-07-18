package goreq

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
	"io/ioutil"
)

var defaultTransport http.RoundTripper = &http.Transport{MaxIdleConns: 10, IdleConnTimeout: 30 * time.Second}
var defaultClient = &http.Client{Transport: defaultTransport}
var defaultOptions = &ReqOptions{
	Headers:        make(http.Header),
	QueryString:    make(url.Values),
}

/*
GoReq stores details that are required to interact with a http request and represents the methods to manipulates the request data.

Generally, you acquire a GoRep by calling goReq.Req() method.

GoReq is not thread safe, so do not use it in concurrency case, but you can reuse it to do different http request,
and it is also the suggested usage
*/
type GoReq struct {
	Options   *ReqOptions
	client    *http.Client
	transport *http.Transport
}

// ReqOptions stores information needed during http request
type ReqOptions struct {
	//http method (default: "GET"), case insensitive
	Method string

	//fully qualified uri or a parsed url object from url.parse()
	Url string

	//http headers (default: {})
	Headers http.Header

	// follow HTTP 3xx responses as redirects (default: true).
	FollowRedirect *bool

	// disable keep alive feature (default: false)
	DisableKeepAlive *bool

	// if not nil, remember cookies for future use (or define your custom cookie jar; see examples section)
	Jar *cookiejar.Jar

	//an HTTP proxy url to be used
	Proxy *string

	// object containing querystring values to be appended to the uri
	QueryString url.Values

	bodyContent httpReqBody

	// e.g 15 * time.Second
	Timeout time.Duration

	HeadersToBeRemove []string
}

// httpReqBody is an interface that providers the request body content and content type
type httpReqBody interface {
	build() (contentType string, data io.Reader)
}

// formContent implements httpReqBody to use form data in http request
type formContent struct {
	content url.Values
}

func (req *GoReq) inToBeRemovedHeader(k string) bool {
	if req.Options.HeadersToBeRemove == nil {
		return false
	}

	for _, key := range req.Options.HeadersToBeRemove {
		if key == k {
			return true
		}
	}

	return false
}

func (req *GoReq) prepareReq() (io.ReadCloser, *http.Response, error) {
	if req.Options.Proxy != nil {
		parsedProxyUrl, err := url.Parse(*req.Options.Proxy)

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
			if !req.inToBeRemovedHeader(k) {
				httpReq.Header[k] = v
			}
		}
	}

	if err != nil {
		return nil, nil, err
	}

	if req.Options.FollowRedirect == nil {
		req.client.CheckRedirect = nil
	} else {
		if *req.Options.FollowRedirect {
			req.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		} else {
			req.client.CheckRedirect = nil
		}
	}

	if req.Options.DisableKeepAlive != nil{
		httpReq.Close = *req.Options.DisableKeepAlive
	}

	resp, err := req.client.Do(httpReq)

	if err != nil {
		return nil, nil, err
	}

	return resp.Body, resp, nil
}

func getStringReader(resp *http.Response, reader io.ReadCloser) (io.ReadCloser, error) {
	if resp.Header.Get("Content-Encoding") == "gzip" {
		defer reader.Close()
		reader, err := gzip.NewReader(resp.Body)

		if err != nil {
			return nil, err
		}

		return reader, nil
	}

	return reader, nil
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

	if copyFrom.FollowRedirect != nil {
		redirect := *copyFrom.FollowRedirect
		copyTo.FollowRedirect = &redirect
	}

	if copyFrom.DisableKeepAlive != nil {
		keepAlive := *copyFrom.DisableKeepAlive
		copyTo.DisableKeepAlive = &keepAlive
	}

	if copyFrom.Proxy != nil {
		proxy := *copyFrom.Proxy
		copyTo.Proxy = &proxy
	}

	if copyTo.Headers == nil {
		copyTo.Headers = make(http.Header)
	}

	for k, v := range copyFrom.Headers {
		copyTo.Headers[k] = v
	}

	return copyTo
}

// Req returns a GoRep with a nother GoRep options
func (req *GoReq) Req(options *ReqOptions) *GoReq {
	mergedOptions := *req.Options
	return Req(mergeOptions(&mergedOptions, options))
}

/*
 Req prepares a GoReq instance. A basic example of using this would be:
   req := goreq.Req(nil)
 */
func Req(options *ReqOptions) *GoReq {
	goReq := GoReq{}
	goReq.transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	goReq.client = &http.Client{
		Transport: goReq.transport,
	}

	if options == nil {
		copyDefault := *defaultOptions
		goReq.Options = &copyDefault
	} else {
		goReq.Options = options
	}

	return &goReq
}

/*
 Post indicates use post method and the post url, equals to:
    req.Options.Url = url
    req.Options.Method = "POST"
 */
func (req *GoReq) Post(url string) *GoReq {
	req.Options.Url = url
	req.Options.Method = "POST"
	return req
}

/*
 Get indicates use get method and the get url, equals to:
    req.Options.Url = url
    req.Options.Method = "GET"
 */
func (req *GoReq) Get(url string) *GoReq {
	req.Options.Url = url
	req.Options.Method = "Get"
	return req
}

/*
 FormData sets the http request body and makes the content-type to "application/x-www-form-urlencoded"
 */
func (req *GoReq) FormData(formData url.Values) *GoReq {
	req.Options.bodyContent = &formContent{
		content: formData,
	}
	return req
}

/*
 JsonString sets the http request body and makes the content-type to "application/json"
 */
func (req *GoReq) JsonString(jsonStr []byte) *GoReq {
	req.Options.bodyContent = &jsonContent{
		content: jsonStr,
	}

	return req
}

/*
 JsonObject sets the http request body and makes the content-type to "application/json"
 */
func (req *GoReq) JsonObject(jsonObj interface{}) *GoReq {
	req.Options.bodyContent = &jsonObjContent{
		content: jsonObj,
	}

	return req
}

/*
 PipeToResponse pipes the result to a http.Response
 */
func (req *GoReq) PipeToResponse(w http.ResponseWriter) error {
	reader, resp, err := req.prepareReq()

	removeResHeaders := map[string]interface{}{
		"Connection": 1,
	}

	if err != nil {
		return err
	}
	defer (reader).Close()

	p := make([]byte, 4)

	for k, v := range resp.Header {
		if removeResHeaders[k] == nil {
			for _, vSub := range v {
				w.Header().Add(k, vSub)
			}
		}
	}

	w.WriteHeader(resp.StatusCode)
	for {
		n, err := (reader).Read(p)

		if err != nil {
			if err == io.EOF {
				w.Write(p[:n])
				break
			}

			return err
		}

		w.Write(p[:n])
	}

	return nil
}

/*
 PipeFromHttpReq pips the http.Request content to goReq request
 */
func (req *GoReq) PipeFromHttpReq(r *http.Request) *GoReq {
	removeReqHeaders := map[string]interface{}{
		"Connection": 1,
		"Referer":    1,
		"Origin":     1,
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

// PipeStream  pipes the response to a writer (e.g: file stream).
func (req *GoReq) PipeStream(writer io.Writer) error {
	reader, resp, err := req.prepareReq()
	readerCloser, err := getStringReader(resp, reader)

	if err != nil {
		return err
	}
	defer (readerCloser).Close()

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

// PipeReq pipes the request result to next request
func (req *GoReq) PipeReq(nextReq *GoReq) (*GoReq, error) {
	reader, resp, err := req.prepareReq()
	readerCloser, err := getStringReader(resp, reader)

	if err != nil {
		return nil, err
	}

	nextReq.Options.bodyContent = &reqPipeContent{reader: readerCloser, contentType: resp.Header.Get("reqPipeContent")}
	return nextReq, nil
}

// UnmarshalJson unmarshals json result to a model
func (req *GoReq) UnmarshalJson(result interface{}) (*http.Response, error) {
	body, resp, err := req.Do()

	if err == nil {
		err = json.Unmarshal((body), result)
	}

	return resp, err
}

// Do method sends the request and return the result
func (req *GoReq) Do() ([]byte, *http.Response, error) {
	reader, resp, err := req.prepareReq()
	readerCloser, err := getStringReader(resp, reader)
	if err != nil {
		return nil, nil, err
	}

	defer (readerCloser).Close()

	body, err := ioutil.ReadAll(readerCloser)

	if err != nil {
		return nil, nil, err
	}

	return body, resp, nil
}
