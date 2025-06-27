// Package mysql 提供MySQL数据库访问功能
// 负责数据库连接管理、SQL查询执行等底层数据操作
package mysql

import (
	"bluebell/setting" // 导入配置包，获取数据库连接配置
	"fmt"              // 导入格式化输出包，用于构建连接字符串

	_ "github.com/go-sql-driver/mysql" // 导入MySQL驱动，下划线表示只执行init函数
	"github.com/jmoiron/sqlx"          // 导入sqlx包，提供更便捷的数据库操作接口
)

// db 全局数据库连接对象
// 使用sqlx.DB类型，提供连接池管理和便捷的查询方法
var db *sqlx.DB

// Init 初始化MySQL数据库连接
// 根据配置信息建立数据库连接，设置连接池参数
// 参数 cfg: MySQL配置信息，包含主机、端口、用户名、密码等
// 返回值: 错误信息，成功时返回nil
func Init(cfg *setting.MySQLConfig) (err error) {
	// ==================== 第一步：构建数据库连接字符串 ====================
	// 格式："user:password@tcp(host:port)/dbname?parseTime=true&loc=Local"
	// parseTime=true: 自动将数据库时间类型转换为Go的time.Time
	// loc=Local: 设置时区为本地时区
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)

	// ==================== 第二步：建立数据库连接 ====================
	// 使用sqlx.Connect建立连接，会自动进行连接测试
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		// 连接失败，返回错误
		return
	}

	// ==================== 第三步：配置连接池参数 ====================
	// 设置最大打开连接数，控制并发连接数量
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	// 设置最大空闲连接数，控制连接池中保持的空闲连接
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	return
}

// Close 关闭MySQL数据库连接
// 程序退出时调用，确保数据库连接正确释放
func Close() {
	// 关闭数据库连接，忽略可能的错误
	_ = db.Close()
}

// SaveVoteData 保存投票数据到MySQL
// 参数 postID: 帖子ID
// 参数 userID: 用户ID
// 参数 voteValue: 投票值（1=赞成，-1=反对）
// 返回值: 错误信息，成功时返回nil
func SaveVoteData(postID, userID int64, voteValue int8) error {
	// 使用REPLACE INTO实现插入或更新
	// 适配现有表结构：post_vote表，字段为post_id, user_id, vote_type
	sqlStr := `REPLACE INTO post_vote(post_id, user_id, vote_type) VALUES(?, ?, ?)`
	_, err := db.Exec(sqlStr, postID, userID, voteValue)
	return err
}
