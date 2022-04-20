package middleware

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"golang.org/x/time/rate"
)
// 访客http访问频率处理中间件
// @author tekintian@gmail.com
// 创建一个自定义visitor结构体，包含每个访问者的限流器和最后一次看到访问者的时间。
type Visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

//用于JSON输出
type JsonResp struct {
	Code int         `json:"code"` //响应编码 0 成功 500 错误 403 无权限
	Msg  string      `json:"msg"`  //消息
	Data interface{} `json:"data"` //数据内容
}

// 更改映射以保存类型visitor的值。
var visitors = make(map[string]*Visitor)
var mu sync.Mutex

// 限制模式 url, ua, 默认 用户IP+UA+URL
var rateMode = "url"

// 运行一个后台goroutine从访客map中删除旧的条目。
func Init(ctx context.Context) {
	gmode, _ := g.Cfg().Get(ctx, "middleware.Visitor.rateMode")
	rateMode = gmode.String()

	// 开启一个协程执行检查和删除访客
	go cleanupVisitors(ctx)
}

func getVisitor(ctx context.Context, visitorkey string) *Visitor {
	mu.Lock()
	defer mu.Unlock()

	visitor, exists := visitors[visitorkey]
	if !exists {
		vconf, _ := g.Cfg().Get(ctx, "middleware.Visitor")
		svmap := vconf.MapStrVar()
		limit := svmap["limit"].Int64()
		burst := svmap["burst"].Int()

		//将int64转换为 time.Dutation
		duration := time.Duration(limit * int64(time.Second))
		//令牌桶容量 3 每秒可以向桶中放1个 token
		limiter := rate.NewLimiter(rate.Every(duration), burst)
		// 在创建新访问者时，添加当前时间。
		visitor := &Visitor{limiter, time.Now()}
		visitors[visitorkey] = visitor
		return visitor
	}
	// 更新访客最后一次出现的时间。
	visitor.lastSeen = time.Now()
	return visitor
}

// 每1分钟检查map上有没有超过 x分钟 的访客，如果有删除。
func cleanupVisitors(ctx context.Context) {
	for {
		time.Sleep(1 * time.Minute)
		mu.Lock()
		//从配置文件获取访客删除间隔
		cleanInterval, _ := g.Cfg().Get(ctx, "middleware.Visitor.cleanInterval")
		duration := time.Duration(cleanInterval.Int64() * int64(time.Minute))
		for visitorkey, v := range visitors {
			// 上次到现在的间隔时间
			if time.Since(v.lastSeen) > duration {
				delete(visitors, visitorkey)
				fmt.Println("visitor key delete!", visitorkey, visitors)
			}
		}
		mu.Unlock()
	}
}

// 访客限流处理
// @author tekintian@gmail.com
func VisitorHandler(r *ghttp.Request) {
	ctx := r.Context()

	visitorKeyStr := r.GetClientIp()
	switch rateMode {
	case "url":
		visitorKeyStr += "#" + r.GetUrl()
	case "ua":
		visitorKeyStr += "#" + r.GetHeader("User-Agent")
	case "ip":
	default:
		visitorKeyStr += ("#" + r.GetHeader("User-Agent") + "#" + r.GetUrl())
	}
	// 区别用户的标识 IP+浏览器UA+请求URL 可配置
	visitorKey := sha1Code(visitorKeyStr)

	visitor := getVisitor(ctx, visitorKey)

	if visitor.limiter.Allow() {
		r.Middleware.Next()
	} else {
		r.Response.WriteJsonExit(JsonResp{
			Code:  509,
			Msg:   g.Cfg().MustGet(ctx, "middleware.Visitor.blockMsg").String(),
		})
	}
}

// sha1加密
func sha1Code(txt string) string {
	o := sha1.New()
	o.Write([]byte(txt))
	return hex.EncodeToString(o.Sum(nil))
}
