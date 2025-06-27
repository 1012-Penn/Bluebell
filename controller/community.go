// Package controller 提供社区相关的HTTP请求处理功能
// 包括获取社区列表、获取社区详情等操作
package controller

import (
	"bluebell/logic" // 导入业务逻辑层，处理社区相关的业务规则
	"strconv"        // 导入字符串转换包，用于类型转换

	"github.com/gin-gonic/gin" // 导入Gin Web框架
	"go.uber.org/zap"          // 导入结构化日志包
)

// ==================== 社区相关功能 ====================

// CommunityHandler 处理获取社区列表请求
// 查询所有社区信息，返回社区ID和社区名称列表
// 用于前端展示社区分类选择
// 参数 c: Gin上下文，包含HTTP请求和响应信息
func CommunityHandler(c *gin.Context) {
	// ==================== 第一步：获取社区列表数据 ====================
	// 调用业务逻辑层获取所有社区信息
	data, err := logic.GetCommunityList()
	if err != nil {
		// 获取失败，记录错误日志
		zap.L().Error("logic.GetCommunityList() failed", zap.Error(err))
		// 不轻易把服务端报错暴露给外面，返回通用服务器繁忙错误
		ResponseError(c, CodeServerBusy)
		return
	}

	// ==================== 第二步：返回社区列表数据 ====================
	ResponseSuccess(c, data)
}

// CommunityDetailHandler 处理获取社区详情请求
// 根据社区ID获取指定社区的详细信息
// 参数 c: Gin上下文，包含HTTP请求和响应信息
func CommunityDetailHandler(c *gin.Context) {
	// ==================== 第一步：参数获取和验证 ====================
	// 从URL路径参数中获取社区ID
	idStr := c.Param("id") // 获取URL参数

	// 将字符串类型的社区ID转换为int64类型
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// 参数转换失败，返回参数错误
		ResponseError(c, CodeInvalidParam)
		return
	}

	// ==================== 第二步：获取社区详情数据 ====================
	// 根据社区ID从数据库获取社区详细信息
	data, err := logic.GetCommunityDetail(id)
	if err != nil {
		// 获取失败，记录错误日志
		zap.L().Error("logic.GetCommunityList() failed", zap.Error(err))
		// 不轻易把服务端报错暴露给外面，返回通用服务器繁忙错误
		ResponseError(c, CodeServerBusy)
		return
	}

	// ==================== 第三步：返回社区详情数据 ====================
	ResponseSuccess(c, data)
}
