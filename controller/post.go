// Package controller 提供帖子相关的HTTP请求处理功能
// 包括创建帖子、获取帖子详情、获取帖子列表等操作
package controller

import (
	"bluebell/logic"  // 导入业务逻辑层，处理帖子相关的业务规则
	"bluebell/models" // 导入数据模型，定义帖子相关的数据结构
	"strconv"         // 导入字符串转换包，用于类型转换

	"github.com/gin-gonic/gin" // 导入Gin Web框架
	"go.uber.org/zap"          // 导入结构化日志包
	// swagger 嵌入文件
)

// CreatePostHandler 创建帖子的处理函数
// 接收客户端发帖请求，验证参数，调用业务逻辑创建帖子
// 参数 c: Gin上下文，包含HTTP请求和响应信息
func CreatePostHandler(c *gin.Context) {
	// ==================== 第一步：参数获取和验证 ====================
	// 创建帖子结构体实例，用于接收JSON请求体数据
	p := new(models.Post)

	// 将JSON请求体绑定到帖子结构体，自动进行参数验证
	if err := c.ShouldBindJSON(p); err != nil {
		// 参数验证失败，记录调试和错误日志
		zap.L().Debug("c.ShouldBindJSON(p) error", zap.Any("err", err))
		zap.L().Error("create post with invalid param")
		ResponseError(c, CodeInvalidParam)
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
	// 设置帖子的作者ID
	p.AuthorID = userID

	// ==================== 第三步：创建帖子 ====================
	// 调用业务逻辑层创建帖子
	if err := logic.CreatePost(p); err != nil {
		// 创建失败，记录错误日志
		zap.L().Error("logic.CreatePost(p) failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// ==================== 第四步：返回成功响应 ====================
	ResponseSuccess(c, nil)
}

// GetPostDetailHandler 获取帖子详情的处理函数
// 根据帖子ID获取帖子的详细信息
// 参数 c: Gin上下文，包含HTTP请求和响应信息
func GetPostDetailHandler(c *gin.Context) {
	// ==================== 第一步：参数获取和验证 ====================
	// 从URL路径参数中获取帖子ID
	pidStr := c.Param("id")

	// 将字符串类型的帖子ID转换为int64类型
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		// 参数转换失败，记录错误日志
		zap.L().Error("get post detail with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// ==================== 第二步：获取帖子数据 ====================
	// 根据帖子ID从数据库获取帖子详细信息
	data, err := logic.GetPostById(pid)
	if err != nil {
		// 获取失败，记录错误日志
		zap.L().Error("logic.GetPostById(pid) failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// ==================== 第三步：返回帖子数据 ====================
	ResponseSuccess(c, data)
}

// GetPostListHandler 获取帖子列表的处理函数（基础版本）
// 获取分页的帖子列表，按默认排序方式
// 参数 c: Gin上下文，包含HTTP请求和响应信息
func GetPostListHandler(c *gin.Context) {
	// ==================== 第一步：获取分页参数 ====================
	// 从请求中获取页码和每页大小
	page, size := getPageInfo(c)

	// ==================== 第二步：获取帖子列表数据 ====================
	// 调用业务逻辑层获取帖子列表
	data, err := logic.GetPostList(page, size)
	if err != nil {
		// 获取失败，记录错误日志
		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// ==================== 第三步：返回帖子列表数据 ====================
	ResponseSuccess(c, data)
}

// GetPostListHandler2 升级版帖子列表接口
// 支持按社区、时间或分数排序查询帖子列表
//
// @Summary 升级版帖子列表接口
// @Description 可按社区按时间或分数排序查询帖子列表接口
// @Tags 帖子相关接口(api分组展示使用的)
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer JWT"
// @Param object query models.ParamPostList false "查询参数"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponsePostList
// @Router /posts2 [get]
//
// 参数 c: Gin上下文，包含HTTP请求和响应信息
func GetPostListHandler2(c *gin.Context) {
	// ==================== 第一步：初始化查询参数 ====================
	// GET请求参数(query string)：/api/v1/posts2?page=1&size=10&order=time
	// 初始化结构体时指定默认参数
	p := &models.ParamPostList{
		Page:  1,                // 默认第1页
		Size:  10,               // 默认每页10条
		Order: models.OrderTime, // 默认按时间排序
	}

	// ==================== 第二步：绑定查询参数 ====================
	// c.ShouldBind() 根据请求的数据类型选择相应的方法去获取数据
	// c.ShouldBindQuery() 专门用于获取URL查询参数
	if err := c.ShouldBindQuery(p); err != nil {
		// 参数绑定失败，记录错误日志
		zap.L().Error("GetPostListHandler2 with invalid params", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// ==================== 第三步：获取帖子列表数据 ====================
	// 调用业务逻辑层获取帖子列表（新版本，支持多种排序方式）
	data, err := logic.GetPostListNew(p)
	if err != nil {
		// 获取失败，记录错误日志
		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// ==================== 第四步：返回帖子列表数据 ====================
	ResponseSuccess(c, data)
}

// 根据社区去查询帖子列表（已注释，保留作为参考）
//func GetCommunityPostListHandler(c *gin.Context) {
//	// 初始化结构体时指定初始参数
//	p := &models.ParamCommunityPostList{
//		ParamPostList: &models.ParamPostList{
//			Page:  1,
//			Size:  10,
//			Order: models.OrderTime,
//		},
//	}
//	//c.ShouldBind()  根据请求的数据类型选择相应的方法去获取数据
//	//c.ShouldBindJSON() 如果请求中携带的是json格式的数据，才能用这个方法获取到数据
//	if err := c.ShouldBindQuery(p); err != nil {
//		zap.L().Error("GetCommunityPostListHandler with invalid params", zap.Error(err))
//		ResponseError(c, CodeInvalidParam)
//		return
//	}
//
//	// 获取数据
//	data, err := logic.GetCommunityPostList(p)
//	if err != nil {
//		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
//		ResponseError(c, CodeServerBusy)
//		return
//	}
//	ResponseSuccess(c, data)
//}
