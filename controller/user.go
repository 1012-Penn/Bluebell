// Package controller 提供HTTP请求处理功能
// 负责接收客户端请求、参数验证、调用业务逻辑层、返回响应
package controller

import (
	"bluebell/dao/mysql" // 导入MySQL数据访问层，用于错误类型判断
	"bluebell/logic"     // 导入业务逻辑层，处理具体的业务规则
	"bluebell/models"    // 导入数据模型，定义请求参数结构
	"errors"             // 导入错误处理包
	"fmt"                // 导入格式化输出包

	"github.com/go-playground/validator/v10" // 导入参数验证器
	"go.uber.org/zap"                        // 导入结构化日志包

	"github.com/gin-gonic/gin" // 导入Gin Web框架
)

// SignUpHandler 处理用户注册请求
// 接收客户端注册请求，验证参数，调用业务逻辑，返回注册结果
// 参数 c: Gin上下文，包含HTTP请求和响应信息
func SignUpHandler(c *gin.Context) {
	// ==================== 第一步：参数获取和验证 ====================
	// 创建注册参数结构体实例
	p := new(models.ParamSignUp)

	// 将JSON请求体绑定到参数结构体，自动进行参数验证
	if err := c.ShouldBindJSON(p); err != nil {
		// 参数验证失败，记录错误日志
		zap.L().Error("SignUp with invalid param", zap.Error(err))

		// 判断错误类型是否为验证器错误
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			// 非验证器错误，返回通用参数错误
			ResponseError(c, CodeInvalidParam)
			return
		}
		// 验证器错误，返回具体的验证错误信息（已翻译为中文）
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}

	// ==================== 第二步：业务逻辑处理 ====================
	// 调用业务逻辑层进行用户注册
	if err := logic.SignUp(p); err != nil {
		// 注册失败，记录错误日志
		zap.L().Error("logic.SignUp failed", zap.Error(err))

		// 判断具体错误类型，返回相应的错误码
		if errors.Is(err, mysql.ErrorUserExist) {
			// 用户已存在错误
			ResponseError(c, CodeUserExist)
			return
		}
		// 其他错误，返回服务器繁忙
		ResponseError(c, CodeServerBusy)
		return
	}

	// ==================== 第三步：返回成功响应 ====================
	ResponseSuccess(c, nil)
}

// LoginHandler 处理用户登录请求
// 接收客户端登录请求，验证用户名密码，生成JWT token，返回登录结果
// 参数 c: Gin上下文，包含HTTP请求和响应信息
func LoginHandler(c *gin.Context) {
	// ==================== 第一步：参数获取和验证 ====================
	// 创建登录参数结构体实例
	p := new(models.ParamLogin)

	// 将JSON请求体绑定到参数结构体，自动进行参数验证
	if err := c.ShouldBindJSON(p); err != nil {
		// 参数验证失败，记录错误日志
		zap.L().Error("Login with invalid param", zap.Error(err))

		// 判断错误类型是否为验证器错误
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			// 非验证器错误，返回通用参数错误
			ResponseError(c, CodeInvalidParam)
			return
		}
		// 验证器错误，返回具体的验证错误信息（已翻译为中文）
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}

	// ==================== 第二步：业务逻辑处理 ====================
	// 调用业务逻辑层进行用户登录验证
	user, err := logic.Login(p)
	if err != nil {
		// 登录失败，记录错误日志（包含用户名信息）
		zap.L().Error("logic.Login failed", zap.String("username", p.Username), zap.Error(err))

		// 判断具体错误类型，返回相应的错误码
		if errors.Is(err, mysql.ErrorUserNotExist) {
			// 用户不存在错误
			ResponseError(c, CodeUserNotExist)
			return
		}
		// 密码错误
		ResponseError(c, CodeInvalidPassword)
		return
	}

	// ==================== 第三步：返回成功响应 ====================
	// 登录成功，返回用户信息和JWT token
	ResponseSuccess(c, gin.H{
		"user_id":   fmt.Sprintf("%d", user.UserID), // 将int64类型的用户ID转换为字符串返回，避免前端精度丢失
		"user_name": user.Username,                  // 返回用户名
		"token":     user.Token,                     // 返回JWT token，用于后续接口认证
	})
}
