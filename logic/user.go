// Package logic 提供业务逻辑处理功能
// 负责处理具体的业务规则，连接控制器层和数据访问层
package logic

import (
	"bluebell/dao/mysql"     // 导入MySQL数据访问层，用于用户数据操作
	"bluebell/models"        // 导入数据模型，定义业务数据结构
	"bluebell/pkg/jwt"       // 导入JWT工具包，用于生成身份令牌
	"bluebell/pkg/snowflake" // 导入雪花算法包，用于生成唯一ID
)

// ==================== 用户业务逻辑处理 ====================

// SignUp 用户注册业务逻辑
// 处理用户注册的完整流程，包括用户存在性检查、ID生成、数据保存等
// 参数 p: 注册参数，包含用户名和密码
// 返回值: 错误信息，成功时返回nil
func SignUp(p *models.ParamSignUp) (err error) {
	// ==================== 第一步：检查用户是否已存在 ====================
	// 调用数据访问层检查用户名是否已被注册
	if err := mysql.CheckUserExist(p.Username); err != nil {
		// 如果用户已存在，返回相应错误
		return err
	}

	// ==================== 第二步：生成用户唯一ID ====================
	// 使用雪花算法生成全局唯一的用户ID
	userID := snowflake.GenID()

	// ==================== 第三步：构造用户数据模型 ====================
	// 创建用户实例，包含生成的ID、用户名和密码
	user := &models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password, // 注意：这里的密码应该已经在前端或控制器层进行了加密
	}

	// ==================== 第四步：保存用户数据到数据库 ====================
	// 调用数据访问层将用户信息插入数据库
	return mysql.InsertUser(user)
}

// Login 用户登录业务逻辑
// 处理用户登录的完整流程，包括密码验证、JWT令牌生成等
// 参数 p: 登录参数，包含用户名和密码
// 返回值: 用户信息（包含JWT令牌）和错误信息
func Login(p *models.ParamLogin) (user *models.User, err error) {
	// ==================== 第一步：构造用户查询对象 ====================
	// 创建用户实例，用于数据库查询
	user = &models.User{
		Username: p.Username,
		Password: p.Password, // 注意：这里的密码应该已经在前端或控制器层进行了加密
	}

	// ==================== 第二步：验证用户登录信息 ====================
	// 调用数据访问层验证用户名和密码
	// 传递的是指针，验证成功后能拿到user.UserID等完整信息
	if err := mysql.Login(user); err != nil {
		// 登录验证失败，返回错误
		return nil, err
	}

	// ==================== 第三步：生成JWT身份令牌 ====================
	// 使用用户ID和用户名生成JWT令牌，用于后续接口的身份认证
	token, err := jwt.GenToken(user.UserID, user.Username)
	if err != nil {
		// JWT生成失败，返回错误
		return
	}

	// ==================== 第四步：设置用户令牌信息 ====================
	// 将生成的JWT令牌保存到用户对象中，返回给客户端
	user.Token = token
	return
}
