// Package controller 提供参数验证和国际化翻译功能
// 基于validator库实现请求参数的自动验证，支持中英文错误信息翻译
package controller

import (
	"bluebell/models" // 导入数据模型，定义请求参数结构
	"fmt"             // 导入格式化输出包
	"reflect"         // 导入反射包，用于获取结构体字段信息
	"strings"         // 导入字符串包，用于字符串处理

	"github.com/gin-gonic/gin/binding"                                      // 导入Gin绑定包，用于修改验证器引擎
	"github.com/go-playground/locales/en"                                   // 导入英文语言包
	"github.com/go-playground/locales/zh"                                   // 导入中文语言包
	ut "github.com/go-playground/universal-translator"                      // 导入通用翻译器
	"github.com/go-playground/validator/v10"                                // 导入验证器库
	enTranslations "github.com/go-playground/validator/v10/translations/en" // 导入英文翻译
	zhTranslations "github.com/go-playground/validator/v10/translations/zh" // 导入中文翻译
)

// trans 全局翻译器实例
// 用于将验证错误信息翻译成指定语言
var trans ut.Translator

// InitTrans 初始化验证器翻译器
// 配置Gin框架的验证器，支持中英文错误信息翻译
// 参数 locale: 语言环境，如"zh"表示中文，"en"表示英文
// 返回值: 错误信息，成功时返回nil
func InitTrans(locale string) (err error) {
	// ==================== 第一步：获取Gin的验证器引擎 ====================
	// 修改gin框架中的Validator引擎属性，实现自定义验证功能
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {

		// ==================== 第二步：注册自定义标签名函数 ====================
		// 注册一个获取json tag的自定义方法
		// 这样验证错误信息会显示JSON字段名而不是Go结构体字段名
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			// 获取json标签的第一个部分（逗号分隔）
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				// 如果json标签是"-"，表示该字段不参与JSON序列化，返回空字符串
				return ""
			}
			return name
		})

		// ==================== 第三步：注册自定义结构体验证函数 ====================
		// 为SignUpParam注册自定义校验方法
		// 用于验证密码和确认密码是否一致
		v.RegisterStructValidation(SignUpParamStructLevelValidation, models.ParamSignUp{})

		// ==================== 第四步：创建语言翻译器 ====================
		zhT := zh.New() // 中文翻译器
		enT := en.New() // 英文翻译器

		// 创建通用翻译器
		// 第一个参数是备用（fallback）的语言环境
		// 后面的参数是应该支持的语言环境（支持多个）
		// uni := ut.New(zhT, zhT) 也是可以的
		uni := ut.New(enT, zhT, enT)

		// ==================== 第五步：获取指定语言的翻译器 ====================
		// locale 通常取决于 http 请求头的 'Accept-Language'
		var ok bool
		// 也可以使用 uni.FindTranslator(...) 传入多个locale进行查找
		trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		// ==================== 第六步：注册翻译器 ====================
		// 根据语言环境注册相应的翻译器
		switch locale {
		case "en":
			// 注册英文翻译器
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		case "zh":
			// 注册中文翻译器
			err = zhTranslations.RegisterDefaultTranslations(v, trans)
		default:
			// 默认使用英文翻译器
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		}
		return
	}
	return
}

// removeTopStruct 去除提示信息中的结构体名称
// 将错误信息中的"结构体名.字段名"格式转换为"字段名"格式
// 参数 fields: 包含结构体名称的错误信息映射
// 返回值: 去除结构体名称后的错误信息映射
func removeTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		// field可能是"SignUpParam.Username",err可能是"Username is required"
		//这个函数就是去掉.之前的部分,只保留.之后的部分,然后赋值给res
		// 找到第一个点号的位置，取点号后面的部分作为字段名
		res[field[strings.Index(field, ".")+1:]] = err
	}
	return res
}

//"SignUpParam.Username": "Username is required" --> "Username is required"

// SignUpParamStructLevelValidation 自定义SignUpParam结构体校验函数
// 用于验证注册参数中的密码和确认密码是否一致
// 参数 sl: 结构体级别的验证器，提供当前验证的结构体信息
func SignUpParamStructLevelValidation(sl validator.StructLevel) {
	// 获取当前正在验证的结构体实例
	su := sl.Current().Interface().(models.ParamSignUp)

	// 验证密码和确认密码是否一致
	if su.Password != su.RePassword {
		// 输出错误提示信息，最后一个参数就是传递的param
		// 使用"eqfield"验证规则，表示必须等于指定字段的值
		sl.ReportError(su.RePassword, "re_password", "RePassword", "eqfield", "password")
	}
}
