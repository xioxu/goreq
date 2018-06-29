package testing

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/xioxu/goreq"
)

func TestPost(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/formdata", func(w http.ResponseWriter, r *http.Request) {
		th.TestHeader(t, r, "Content-Type", "application/x-www-form-urlencoded")
		th.TestBody(t, r, "pwd=111&userName=nxu")
		w.WriteHeader(http.StatusOK)
	})

	th.Mux.HandleFunc("/jsongstr", func(w http.ResponseWriter, r *http.Request) {
		th.TestHeader(t, r, "Content-Type", "application/json")
		th.TestBody(t, r, "{ok:1}")
		w.WriteHeader(http.StatusOK)
	})

	th.Mux.HandleFunc("/jsonobj", func(w http.ResponseWriter, r *http.Request) {
		th.TestHeader(t, r, "Content-Type", "application/json")
		th.TestBody(t, r, `{"ok":"abc"}`)
		w.WriteHeader(http.StatusOK)
	})

	th.Mux.HandleFunc("/jsonobj2", func(w http.ResponseWriter, r *http.Request) {
		th.TestHeader(t, r, "Content-Type", "application/json")
		th.TestBody(t, r, `{"Name":"xdw","Age":30}`)
		w.WriteHeader(http.StatusOK)
	})

	req := goreq.Req(nil)

	postFormData := url.Values{}
	postFormData.Add("userName", "nxu")
	postFormData.Add("pwd", "111")
	req.Post(th.Endpoint() + "formdata").FormData(postFormData).Do()
	req.Post(th.Endpoint() + "jsongstr").JsonString([]byte("{ok:1}")).Do()

	jsonData := map[string]string{
		"ok": "abc",
	}

	req.Post(th.Endpoint() + "jsonobj").JsonObject(jsonData).Do()

	var postData struct {
		Name string
		Age  int
	}

	postData.Name = "xdw"
	postData.Age = 30
	req.Post(th.Endpoint() + "jsonobj2").JsonObject(postData).Do()
}

func TestPipeStream(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/req", func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprintf(w, "abc")
		w.WriteHeader(http.StatusOK)
	})

	req := goreq.Req(nil)
	var writer bytes.Buffer
	req.Get(th.Endpoint() + "req").PipeStream(io.Writer(&writer))
	output := writer.String()
	th.AssertEquals(t, "abc", output)
}

func TestPipeReq(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/req1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "abc")
		w.WriteHeader(http.StatusOK)
	})

	th.Mux.HandleFunc("/req2", func(w http.ResponseWriter, r *http.Request) {

		th.TestBody(t, r, `abc`)
		w.WriteHeader(http.StatusOK)
	})

	req := goreq.Req(nil)
	nextReq, _ := req.Get(th.Endpoint() + "req1").PipeReq(goreq.Req(nil).Post(th.Endpoint() + "req2"))
	nextReq.Do()
}

func TestGet(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "abc")
	})

	req := goreq.Req(nil)
	body, resp, _ := req.Get(th.Endpoint()).Do()

	th.AssertByteArrayEquals(t, []byte("abc"), body)
	th.AssertEquals(t, 200, resp.StatusCode)
}

func TestPipeFromReq(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/req1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "abc")

		req := goreq.Req(nil)
		req.Get(th.Endpoint() + "req2").PipeFromReq(r).Do()
	})

	th.Mux.HandleFunc("/req2", func(w http.ResponseWriter, r *http.Request) {
		th.TestHeader(t, r, "header1", "headver1_val")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "abc")
	})

	req := goreq.Req(&goreq.ReqOptions{Headers: map[string][]string{
		"header1": []string{"headver1_val"},
	}})
	req.Get(th.Endpoint() + "req1").Do()
}

func TestRemoveHeader(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/req1", func(w http.ResponseWriter, r *http.Request) {
		th.TestHeader(t, r, "header1", "headver1_val")
		th.TestHeader(t, r, "header2", "")

		w.WriteHeader(http.StatusOK)
	})

	req := goreq.Req(&goreq.ReqOptions{Headers: map[string][]string{
		"header1": {"headver1_val"},
		"header2": {"headver2_val"},
	}, HeadersToBeRemove: []string{"header2"}})
	req.Get(th.Endpoint() + "req1").Do()
}
func TestGlobalOptions(t *testing.T) {
	req := goreq.Req(&goreq.ReqOptions{FollowRedirect: &goreq.NullableBool{Value: true}, Proxy: &goreq.NullableString{Value: "http://localhost:8888"}})
	req1 := req.Req(nil).Get("http://www.abc.com")
	req2 := req.Req(&goreq.ReqOptions{FollowRedirect: &goreq.NullableBool{Value: false}}).Get("http://www.abc.com")

	th.AssertEquals(t, "http://localhost:8888", req1.Options.Proxy.Value)
	th.AssertEquals(t, "http://localhost:8888", req2.Options.Proxy.Value)
	th.AssertEquals(t, false, req2.Options.FollowRedirect.Value)
	th.AssertEquals(t, true, req.Options.FollowRedirect.Value)
}
