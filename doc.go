/*
Package goreq provides the options and the members for a simpilest http client library
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
*/

package goreq
