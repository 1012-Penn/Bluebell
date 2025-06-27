// Package controller 提供投票相关的HTTP请求处理功能
// 包括帖子投票、获取投票结果等操作
package controller

import (
	"bluebell/logic"  // 导入业务逻辑层，处理投票相关的业务规则
	"bluebell/models" // 导入数据模型，定义投票相关的数据结构

	"go.uber.org/zap" // 导入结构化日志包

	"github.com/go-playground/validator/v10" // 导入参数验证器

	"github.com/gin-gonic/gin" // 导入Gin Web框架
)

// 投票功能相关注释

// VoteData 投票数据结构（已注释，保留作为参考）
//type VoteData struct {
//	// UserID 从请求中获取当前的用户
//	PostID    int64 `json:"post_id,string"`   // 贴子id
//	Direction int   `json:"direction,string"` // 赞成票(1)还是反对票(-1)
//}

// PostVoteController 处理帖子投票请求
// 接收客户端投票请求，验证参数，调用业务逻辑处理投票
// 支持赞成票(1)和反对票(-1)两种投票方式
// 参数 c: Gin上下文，包含HTTP请求和响应信息
func PostVoteController(c *gin.Context) {
	// ==================== 第一步：参数获取和验证 ====================
	// 创建投票参数结构体实例，用于接收JSON请求体数据
	p := new(models.ParamVoteData)

	// 将JSON请求体绑定到投票参数结构体，自动进行参数验证
	if err := c.ShouldBindJSON(p); err != nil {
		// 判断错误类型是否为验证器错误
		errs, ok := err.(validator.ValidationErrors) // 类型断言
		if !ok {
			// 非验证器错误，返回通用参数错误
			ResponseError(c, CodeInvalidParam)
			return
		}
		// 验证器错误，翻译并去除掉错误提示中的结构体标识，返回具体的验证错误信息
		errData := removeTopStruct(errs.Translate(trans))
		ResponseErrorWithMsg(c, CodeInvalidParam, errData)
		return
	}

	// ==================== 第二步：获取当前用户信息 ====================
	// 从JWT token中获取当前登录用户的ID
	userID, err := getCurrentUserID(c)
	if err != nil {
		// 获取用户ID失败，说明用户未登录或token无效
		ResponseError(c, CodeNeedLogin)
		return
	}

	// ==================== 第三步：处理投票业务逻辑 ====================
	// 调用业务逻辑层处理具体的投票操作
	// 包括：验证帖子是否存在、检查用户是否已投票、更新投票记录、更新帖子分数等
	if err := logic.VoteForPost(userID, p); err != nil {
		// 投票失败，记录错误日志
		zap.L().Error("logic.VoteForPost() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// ==================== 第四步：返回成功响应 ====================
	ResponseSuccess(c, nil)
}
