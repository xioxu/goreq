package goreq

import "fmt"

func ExampleGoReq_Req() {
	req := Req(nil)
	body,_,_ := req.Get("https://www.baidu.com").Do()
	fmt.Print(string(body))
}

