// Package redis 提供Redis缓存数据库访问功能
// 负责Redis连接管理、缓存操作等底层数据访问
package redis

import (
	"context" // 导入上下文包，用于Redis操作超时控制
	"fmt"     // 导入格式化输出包，用于构建连接地址

	"github.com/go-redis/redis/v8" // 导入Redis客户端包，提供Redis操作接口

	"bluebell/setting" // 导入配置包，获取Redis连接配置
)

// 实际生产环境下 context.Background() 按需替换
// 可以根据业务需求使用带超时的context或请求级别的context

// 全局变量定义
var (
	client *redis.Client // Redis客户端实例，提供连接池和操作接口
	Nil    = redis.Nil   // Redis空值常量，用于判断键是否存在
)

// Init 初始化Redis连接
// 根据配置信息建立Redis连接，设置连接池参数
// 参数 cfg: Redis配置信息，包含主机、端口、密码、数据库等
// 返回值: 错误信息，成功时返回nil
func Init(cfg *setting.RedisConfig) (err error) {
	// ==================== 第一步：创建Redis客户端 ====================
	// 使用redis.NewClient创建客户端实例，配置连接参数
	client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), // 构建连接地址,动态配置,根据配置文件中的host和port来构建连接地址
		Password:     cfg.Password,                             // Redis密码，无密码时为空字符串
		DB:           cfg.DB,                                   // 使用的数据库编号，默认使用0号数据库
		PoolSize:     cfg.PoolSize,                             // 连接池大小，控制最大连接数
		MinIdleConns: cfg.MinIdleConns,                         // 最小空闲连接数，保持连接池中的最小连接
	})

	// ==================== 第二步：测试连接 ====================
	// 使用Ping命令测试Redis连接是否正常
	// context.Background() 提供无超时的上下文，生产环境建议使用带超时的context
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		// 连接测试失败，返回错误
		return err
	}

	return nil
}

// Close 关闭Redis连接
// 程序退出时调用，确保Redis连接正确释放
func Close() {
	// 关闭Redis客户端连接，忽略可能的错误
	_ = client.Close()
}
