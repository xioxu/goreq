package testing

import (
	"github.com/xioxu/goreq"
	"net/url"
	"fmt"
	"testing"
)

func TestPost(t *testing.T)  {
	req := goreq.Req(nil)

	postFormData := url.Values{}
	postFormData.Add("userName","nxu")
	postFormData.Add("pwd","111")
	body ,_ , err := req.Post("https://www.baidu.com").FormData(postFormData).Do()

	if err != nil{
		fmt.Print(err)
	} else {
		fmt.Print(string(body))
	}
}

func TestGet(t *testing.T)  {
	req := goreq.Req(nil)
	body,resp,_ := req.Get("https://www.baidu.com").Do()
	fmt.Println(string(body))
	fmt.Println(resp.StatusCode)
}
