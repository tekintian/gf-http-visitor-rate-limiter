# goframe v2框架 http访客频率限制中间件

gf v2 框架使用的访客限流中间件, 基于谷歌令牌桶算法实现! 可自定义限制默认, 访问频率等, 轻量简单,够用就好!

## 配置

在你的gf框架配置文件config.yaml 中增加以下内容
~~~yml

# 中间件配置
middleware:
  # 访客访问频率限制中间件, 作用, 限制单用户访问页面的频率
  Visitor:
    rateMode: all # 限制模式, ip + xxx , 可选 all url , ua 不配置 默认 ip+ua+url
    limit: 1 # 每X秒生成一个访问令牌
    burst: 3 # 令牌桶容量, 即用户 limit秒内可访问的次数
    cleanInterval: 5 # 非活动用户删除间隔 单位: 分钟
~~~

## 初始化/绑定中间件

- 安装依赖
~~~sh
got get github.com/tekintian/gf-http-visitor-rate-limiter/middleware
~~~

- 使用示例
~~~go
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

		// other route

	})

}

~~~

## 测试
完成上面2步后 gf-http-visitor-rate-limiter 中间件就已经成功启用了, 浏览器打开你的页面快速刷新看看效果 :)


