package main

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/tekintian/gf-http-visitor-rate-limiter/middleware"
	// other
)

func main() {
	ctx := context.TODO()

	s := g.Server("yourservername")

	//初始化访客频率处理中间件 注意这个必须在路由注册之前
	middleware.Init(ctx)

	s.Group("/", func(group *ghttp.RouterGroup) {

		//register middleware
		group.Middleware(middleware.VisitorHandler) //访客限流处理中间件

		// test route
		group.ALL("/hello", func(r *ghttp.Request) {
			r.Response.WriteJson(middleware.JsonResp{Code: 200, Msg: "OK", Data: "Hello world gf-http-visitor-rate-limiter !"})
		})
	})

}
