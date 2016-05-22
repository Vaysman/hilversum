package httpproxy

import (
	"github.com/spf13/viper"
	"github.com/elazarl/goproxy"
	jww "github.com/spf13/jwalterweatherman"
	"net/http"
)

func Run(viper *viper.Viper) {
	jww.INFO.Println("Start HTTP rewriter")
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.OnResponse(goproxy.DstHostIs("httpbin.org")).DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		jww.INFO.Printf("Response %v\tContext %v", resp, ctx)
		return resp
	})
	proxy.OnRequest().HijackConnect()
	jww.FATAL.Fatal(http.ListenAndServe(":8080", proxy))
}