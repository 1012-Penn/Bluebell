// Package controller 提供错误码定义和错误信息管理功能
// 统一管理系统中所有的错误码和对应的错误信息
package controller

// ResCode 响应码类型
// 使用int64类型定义，确保有足够的范围存储各种错误码
type ResCode int64

// 系统错误码常量定义
// 使用iota自动递增，从1000开始，避免与HTTP状态码冲突
const (
	CodeSuccess         ResCode = 1000 + iota // 成功码：1000
	CodeInvalidParam                          // 参数错误：1001
	CodeUserExist                             // 用户已存在：1002
	CodeUserNotExist                          // 用户不存在：1003
	CodeInvalidPassword                       // 密码错误：1004
	CodeServerBusy                            // 服务器繁忙：1005

	CodeNeedLogin    // 需要登录：1006
	CodeInvalidToken // 无效token：1007
)

// codeMsgMap 错误码与错误信息的映射表
// 将每个错误码映射到对应的中文错误信息
var codeMsgMap = map[ResCode]string{
	CodeSuccess:         "success",  // 成功
	CodeInvalidParam:    "请求参数错误",   // 请求参数格式或内容有误
	CodeUserExist:       "用户名已存在",   // 注册时用户名已被使用
	CodeUserNotExist:    "用户名不存在",   // 登录时用户名不存在
	CodeInvalidPassword: "用户名或密码错误", // 登录时密码验证失败
	CodeServerBusy:      "服务繁忙",     // 服务器内部错误或暂时不可用

	CodeNeedLogin:    "需要登录",     // 访问需要认证的接口时未提供有效token
	CodeInvalidToken: "无效的token", // JWT token格式错误或已过期
}

// Msg 获取错误码对应的错误信息
// 如果错误码不存在于映射表中，则返回服务器繁忙的错误信息
// 返回值: 错误码对应的中文错误信息
func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		// 如果错误码不存在，返回默认的服务器繁忙信息
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
