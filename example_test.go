package goreq

import (
	"fmt"
	"net/http"
)

func ExampleGoReq_Req() {
	req := Req(&ReqOptions{Proxy:NewString("http://localhost:8888")})

	// req1 still keeps the options of req
	req1 := req.Req(nil)
	req1.Get("http://www.baidu.com").Do()
}

func ExampleReq() {
	req := Req(nil)
	body,_,_ := req.Get("https://www.baidu.com").Do()
	fmt.Print(string(body))
}

func ExampleGoReq_PipeReq() {
	if err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		req := Req(&ReqOptions{
			Method: r.Method,
			Url:    "https://www.baidu.com" + r.RequestURI,
		})

		req.PipeFromHttpReq(r).PipeToResponse(w)
	})); err != nil {
		panic(err)
	}
}

