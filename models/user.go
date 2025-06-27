// Package models 提供数据模型定义功能
// 定义系统中各种业务实体的数据结构，用于数据存储和传输
package models

// User 用户数据模型
// 定义用户的基本信息结构，对应数据库中的用户表
type User struct {
	UserID   int64  `db:"user_id"`  // 用户ID，使用雪花算法生成的唯一标识
	Username string `db:"username"` // 用户名，用于登录和显示
	Password string `db:"password"` // 密码，存储加密后的密码哈希值
	Token    string // JWT令牌，用于身份认证（不存储到数据库）
}
