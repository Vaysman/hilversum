package httpproxy

import (
	"github.com/spf13/viper"
	"github.com/elazarl/goproxy"
	jww "github.com/spf13/jwalterweatherman"
	"net/http"
	"io/ioutil"
	"bytes"
	"regexp"
)

func Run(viper *viper.Viper) {
	urls := map[string]string{
		"Accept-Encoding": "_Accept-Encoding_",
		"Host": "_Host_",
	}

	jww.INFO.Println("Start HTTP rewriter")
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*\\.org:.*$"))).HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		r.Header.Set("X-GoProxy", "yxorPoG-X")
		return r, nil
	})
	proxy.OnResponse(goproxy.DstHostIs("httpbin.org")).DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		jww.INFO.Printf("Response %v\tContext %v", resp, ctx)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			jww.ERROR.Printf("Error: %v URL: %v", err, ctx.Req.URL)
			return goproxy.TextResponse(ctx.Req, "Error getting body")
		}
		err = resp.Body.Close()
		if err != nil {
			jww.ERROR.Printf("Error closing %v: %v", ctx.Req.URL, err)
			return goproxy.TextResponse(ctx.Req, "Error clossing body")
		}
		for old, new := range urls {
			b = bytes.Replace(b, []byte(old), []byte(new), -1)
		}
		body := ioutil.NopCloser(bytes.NewReader(b))
		resp.Body = body

		return resp
	})

	jww.FATAL.Fatal(http.ListenAndServe(":8080", proxy))
}