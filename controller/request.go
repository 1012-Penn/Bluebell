package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

const CtxUserIDKey = "userID"

var ErrorUserNotLogin = errors.New("用户未登录")

// getCurrentUserID 获取当前登录的用户ID
func getCurrentUserID(c *gin.Context) (userID int64, err error) {
	uid, ok := c.Get(CtxUserIDKey)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	userID, ok = uid.(int64)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	return
}

// getPageInfo 获取分页参数
// 从URL查询参数中获取页码(page)和每页大小(size)
// 如果参数无效或缺失，使用默认值
// 参数 c: Gin上下文，包含HTTP请求信息
// 返回值: (页码, 每页大小)
func getPageInfo(c *gin.Context) (int64, int64) {
	// ==================== 第一步：从URL查询参数获取分页信息 ====================
	// c.Query("page") 获取URL中的page参数，例如：/api/posts?page=2&size=20
	// 如果URL中没有page参数，pageStr为空字符串
	// c.Query() 获取URL中的参数，如果参数不存在，返回空字符串
	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	// ==================== 第二步：声明变量 ====================
	// 使用var关键字同时声明多个变量
	// 这些变量会被初始化为零值（int64的零值是0）
	var (
		page int64 // 页码，默认为0
		size int64 // 每页大小，默认为0
		err  error // 错误信息，默认为nil
	)

	// ==================== 第三步：解析页码参数 ====================
	// strconv.ParseInt() 将字符串转换为int64类型
	// 参数说明：
	// - pageStr: 要转换的字符串
	// - 10: 进制（10进制）
	// - 64: 位数（64位整数）
	page, err = strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		// 如果转换失败（比如pageStr为空、不是数字等），使用默认值
		page = 1 // 默认第1页
	}

	// ==================== 第四步：解析每页大小参数 ====================
	// 同样将size字符串转换为int64
	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		// 如果转换失败，使用默认值
		size = 10 // 默认每页10条数据
	}

	// ==================== 第五步：返回解析结果 ====================
	return page, size
}
