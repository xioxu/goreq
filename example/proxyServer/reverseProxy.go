package reverseProxy

import (
	"net/http"
	"github.com/xioxu/goreq"
)

func main() {
	if err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		req := goreq.Req(&goreq.ReqOptions{
			Method: r.Method,
			Url:    "https://www.baidu.com" + r.RequestURI,
		})


		req.PipeFromReq(r).PipeToResponse(w)

	})); err != nil {
		panic(err)
	}
}
