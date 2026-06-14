package routers

import (
	"net/http"

	"github.com/beego/beego/context"
)

func CorsFilter(ctx *context.Context) {
	origin := ctx.Request.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", origin)
	ctx.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, PATCH, OPTIONS")
	ctx.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	ctx.ResponseWriter.Header().Set("Access-Control-Allow-Credentials", "true")
	if ctx.Request.Method == http.MethodOptions {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	}
}
