/*
Package goreq provides the ability to do http request with the simplest code

Example to Request a site

    req := goreq.Req(nil)
	body,_,_ := req.Get("https://www.baidu.com").Do()
	fmt.Print(string(body))

Example to submit a request

	req := goreq.Req(nil)
	postFormData := url.Values{}
	postFormData.Add("userName", "nxu")
	postFormData.Add("pwd", "111")

	body,_,_ := req.Post("https://www.baidu.com").FormData(postFormData).Do()
	fmt.Print(string(body))

This top-level package contains utility functions and data types that are used
throughout the http requesting.
*/
package goreq
