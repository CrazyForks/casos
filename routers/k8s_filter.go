package routers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/beego/beego/context"
	"github.com/beego/beego/logs"
)

func K8sProxyFilter(apiserverOrigin string) func(*context.Context) {
	u, err := url.Parse(apiserverOrigin)
	if err != nil {
		panic("k8s proxy: invalid apiserver origin: " + err.Error())
	}
	rp := httputil.NewSingleHostReverseProxy(u)
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logs.Warning("k8s proxy %s: %v", r.URL.Path, err)
		http.Error(w, "apiserver unavailable", http.StatusBadGateway)
	}

	return func(ctx *context.Context) {
		r := ctx.Request
		// Strip /k8s prefix so apiserver receives /api/v1/... unchanged.
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/k8s")
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		r.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, "/k8s")
		rp.ServeHTTP(ctx.ResponseWriter, r)
	}
}
