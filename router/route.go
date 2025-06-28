// Package router 提供路由配置功能
// 负责设置和管理整个Web应用的路由规则
package router

import (
	"bluebell/controller"  // 导入控制器包，处理具体的业务逻辑
	"bluebell/logger"      // 导入日志包，用于记录应用日志
	"bluebell/middlewares" // 导入中间件包，提供认证等功能
	"net/http"             // 导入HTTP包，提供HTTP状态码等常量
	"time"                 // 导入time包，用于时间处理

	ginSwagger "github.com/swaggo/gin-swagger"   // 导入Swagger文档生成器
	"github.com/swaggo/gin-swagger/swaggerFiles" // 导入Swagger静态文件处理器

	_ "bluebell/docs" // 导入API文档，下划线表示只执行init函数

	"github.com/gin-contrib/pprof" // 导入性能分析工具
	"github.com/gin-gonic/gin"     // 导入Gin Web框架
)

// SetupRouter 设置并返回Gin路由引擎
// 参数 mode: 运行模式（开发模式或发布模式）
// 返回值: 配置好的Gin引擎实例
func SetupRouter(mode string) *gin.Engine {
	// 如果是发布模式，设置Gin为发布模式（减少日志输出，提高性能）
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) // gin设置成发布模式
	}

	// 创建新的Gin引擎实例 也可以用gin.Default()自动包含Logger和Recovery中间件
	r := gin.New()

	// 注册全局中间件
	// GinLogger(): 记录HTTP请求日志
	// GinRecovery(true): 从panic中恢复，避免程序崩溃
	// RateLimitMiddleware: 基于令牌桶算法的限流中间件
	// 参数：每100ms填充一个令牌，令牌桶容量为100
	// 这意味着：QPS限制为10，突发处理能力为100
	r.Use(logger.GinLogger(), logger.GinRecovery(true), middlewares.RateLimitMiddleware(100*time.Millisecond, 100))

	// 健康检查接口 - 用于检测服务是否正常运行
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Swagger API文档接口 - 提供API文档访问
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 创建API v1版本的路由组
	v1 := r.Group("/api/v1")

	// ==================== 无需认证的公开接口 ====================

	// 用户注册接口
	v1.POST("/signup", controller.SignUpHandler)
	// 用户登录接口
	v1.POST("/login", controller.LoginHandler)

	// 获取帖子列表接口（支持按时间排序）
	v1.GET("/posts2", controller.GetPostListHandler2)
	// 获取帖子列表接口（支持按分数排序）
	v1.GET("/posts", controller.GetPostListHandler)
	// 获取社区列表接口
	v1.GET("/community", controller.CommunityHandler)
	// 获取指定社区详情接口
	v1.GET("/community/:id", controller.CommunityDetailHandler)
	// 获取指定帖子详情接口
	v1.GET("/post/:id", controller.GetPostDetailHandler)

	// ==================== 需要JWT认证的接口 ====================

	// 为v1路由组应用JWT认证中间件
	// 此中间件会验证请求头中的JWT token
	v1.Use(middlewares.JWTAuthMiddleware()) // 应用JWT认证中间件

	{
		// 创建新帖子接口（需要登录）
		v1.POST("/post", controller.CreatePostHandler)

		// 投票接口（需要登录）
		v1.POST("/vote", controller.PostVoteController)
	}

	// 注册性能分析工具的路由
	// 可以通过 /debug/pprof/ 访问性能分析数据
	pprof.Register(r) // 注册pprof相关路由

	// 404处理 - 当访问不存在的路由时返回404响应
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})

	// 返回配置好的路由引擎
	return r
}
