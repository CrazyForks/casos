package routers

import (
	"github.com/beego/beego"
	"github.com/beego/beego/context"
)

func ApiFilter(ctx *context.Context) {
	if beego.AppConfig.DefaultBool("isDemoMode", false) && !isAllowedInDemoMode(ctx.Request.Method, ctx.Request.URL.Path) {
		denyRequest(ctx)
	}
}

func isAllowedInDemoMode(method, urlPath string) bool {
	if method == "POST" {
		return urlPath == "/api/signin" || urlPath == "/api/signout"
	}
	return true
}
