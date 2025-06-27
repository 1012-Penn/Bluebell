// Package middlewares 提供HTTP中间件功能
// 包括认证、限流、日志等中间件，用于处理HTTP请求的通用逻辑
package middlewares

import (
	"net/http" // 导入HTTP包，提供HTTP状态码等常量
	"time"     // 导入时间包，用于定义时间间隔

	"github.com/gin-gonic/gin"  // 导入Gin Web框架
	"github.com/juju/ratelimit" // 导入限流工具包，提供令牌桶算法实现
)

// RateLimitMiddleware 基于令牌桶算法的限流中间件
// 使用令牌桶算法控制请求频率，防止系统过载
// 参数 fillInterval: 令牌填充间隔，如2*time.Second表示每2秒填充一个令牌
// 参数 cap: 令牌桶容量，即最大可存储的令牌数量
// 返回值: Gin中间件函数，用于限制请求频率
func RateLimitMiddleware(fillInterval time.Duration, cap int64) func(c *gin.Context) {
	// 创建令牌桶实例
	// fillInterval: 令牌填充间隔，控制令牌生成速率
	// cap: 令牌桶容量，控制突发请求的处理能力
	bucket := ratelimit.NewBucket(fillInterval, cap)

	return func(c *gin.Context) {
		// ==================== 第一步：尝试获取令牌 ====================
		// 尝试从令牌桶中获取一个令牌
		// TakeAvailable(1) 返回实际获取到的令牌数量
		// 如果返回1，说明成功获取到令牌；如果返回0，说明令牌不足
		if bucket.TakeAvailable(1) != 1 {
			// 取不到令牌，说明请求频率过高，返回限流响应
			c.String(http.StatusOK, "rate limit...")
			c.Abort() // 终止后续中间件和处理器执行
			return
		}

		// ==================== 第二步：放行请求 ====================
		// 成功获取到令牌，继续执行后续的中间件和请求处理函数
		c.Next()
	}
}
