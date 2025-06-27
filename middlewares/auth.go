// Package middlewares 提供HTTP中间件功能
// 包括认证、限流、日志等中间件，用于处理HTTP请求的通用逻辑
package middlewares

import (
	"bluebell/controller" // 导入控制器包，用于返回统一格式的错误响应
	"bluebell/pkg/jwt"    // 导入JWT工具包，用于解析和验证token
	"strings"             // 导入字符串处理包，用于分割token字符串

	"github.com/gin-gonic/gin" // 导入Gin Web框架
)

// JWTAuthMiddleware 基于JWT的认证中间件
// 验证请求中的JWT token，提取用户信息并保存到请求上下文中
// 返回值: Gin中间件函数，用于验证JWT token的有效性
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// ==================== 第一步：获取Authorization请求头 ====================
		// 客户端携带Token有三种方式：
		// 1. 放在请求头 (Authorization: Bearer xxx.xxx.xxx)
		// 2. 放在请求体 (JSON中的token字段)
		// 3. 放在URI (查询参数)
		// 这里采用标准的Bearer Token方式，从Authorization请求头获取
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			// 请求头中没有Authorization字段，返回需要登录错误
			controller.ResponseError(c, controller.CodeNeedLogin)
			c.Abort() // 终止后续中间件和处理器执行
			return
		}

		// ==================== 第二步：验证Token格式 ====================
		// 按空格分割Authorization头，格式应为："Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			// Token格式不正确，返回无效token错误
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort() // 终止后续中间件和处理器执行
			return
		}

		// ==================== 第三步：解析和验证JWT Token ====================
		// parts[1]是获取到的tokenString，使用JWT工具包解析它
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			// Token解析失败（可能是格式错误、签名无效、已过期等），返回无效token错误
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort() // 终止后续中间件和处理器执行
			return
		}

		// ==================== 第四步：保存用户信息到请求上下文 ====================
		// 将当前请求的用户ID信息保存到请求的上下文c上
		// 后续的处理函数中可以通过c.Get(controller.CtxUserIDKey)来获取当前请求的用户信息
		c.Set(controller.CtxUserIDKey, mc.UserID)

		// 继续执行后续的中间件和请求处理函数
		c.Next()
	}
}
