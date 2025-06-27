// Package controller 提供HTTP响应处理功能
// 负责统一格式化API响应数据，确保返回格式的一致性
package controller

import (
	"net/http" // 导入HTTP包，提供HTTP状态码等常量

	"github.com/gin-gonic/gin" // 导入Gin Web框架
)

/*
统一响应格式说明：
{
	"code": 10000, // 程序中的错误码，用于标识不同的错误类型
	"msg": xx,     // 提示信息，可以是错误描述或成功提示
	"data": {},    // 数据，成功时返回具体数据，失败时为null
}
*/

// ResponseData 统一响应数据结构
// 所有API接口都使用此结构返回数据，确保前端处理的一致性
type ResponseData struct {
	Code ResCode     `json:"code"`           // 响应码，标识请求处理结果
	Msg  interface{} `json:"msg"`            // 响应消息，可以是字符串或错误详情
	Data interface{} `json:"data,omitempty"` // 响应数据，omitempty表示空值时省略此字段
}

// ResponseError 返回错误响应
// 使用预定义的错误码返回标准错误信息
// 参数 c: Gin上下文
// 参数 code: 预定义的错误码
// 业务逻辑错误统一使用200状态码, 前端根据code来判断是否是业务逻辑错误
func ResponseError(c *gin.Context, code ResCode) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: code,       // 设置错误码
		Msg:  code.Msg(), // 获取错误码对应的标准错误信息
		Data: nil,        // 错误时数据为空
	})
}

// ResponseErrorWithMsg 返回自定义错误响应
// 使用预定义的错误码但返回自定义错误信息
// 参数 c: Gin上下文
// 参数 code: 预定义的错误码
// 参数 msg: 自定义错误信息
func ResponseErrorWithMsg(c *gin.Context, code ResCode, msg interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: code, // 设置错误码
		Msg:  msg,  // 使用自定义错误信息
		Data: nil,  // 错误时数据为空
	})
}

// ResponseSuccess 返回成功响应
// 返回成功状态码和具体的数据内容
// 参数 c: Gin上下文
// 参数 data: 要返回的数据内容
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: CodeSuccess,       // 设置成功码
		Msg:  CodeSuccess.Msg(), // 获取成功码对应的标准成功信息
		Data: data,              // 返回具体的数据内容
	})
}
