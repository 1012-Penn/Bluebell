// Package main 是bluebell项目的入口包
// 负责项目的启动、配置加载、组件初始化等核心功能
package main

import (
	"bluebell/controller"    // 导入控制器包，处理HTTP请求
	"bluebell/dao/mysql"     // 导入MySQL数据访问层
	"bluebell/dao/redis"     // 导入Redis数据访问层
	"bluebell/logger"        // 导入日志包
	"bluebell/pkg/snowflake" // 导入雪花算法包，用于生成唯一ID
	"bluebell/router"        // 导入路由包
	"bluebell/setting"       // 导入配置包
	"fmt"                    // 导入格式化输出包
	"os"                     // 导入操作系统接口包
)

// @title bluebell项目接口文档
// @version 1.0
// @description Go web开发进阶项目实战课程bluebell

// @contact.name liwenzhou
// @contact.url http://www.liwenzhou.com

// @host 127.0.0.1:8084
// @BasePath /api/v1

// main 函数是程序的入口点
// 负责初始化所有必要的组件并启动Web服务器
func main() {
	// 检查命令行参数，确保提供了配置文件路径
	if len(os.Args) < 2 {
		fmt.Println("need config file.eg: bluebell config.yaml")
		return
	}

	// ==================== 第一步：加载配置文件 ====================
	// 从命令行参数指定的配置文件路径加载配置
	if err := setting.Init(os.Args[1]); err != nil {
		fmt.Printf("load config failed, err:%v\n", err)
		return
	}

	// ==================== 第二步：初始化日志系统 ====================
	// 根据配置初始化日志记录器，支持不同级别的日志输出
	if err := logger.Init(setting.Conf.LogConfig, setting.Conf.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}

	// ==================== 第三步：初始化MySQL数据库连接 ====================
	// 建立与MySQL数据库的连接，用于持久化数据存储
	if err := mysql.Init(setting.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	defer mysql.Close() // 程序退出时关闭数据库连接，确保资源释放

	// ==================== 第四步：初始化Redis缓存连接 ====================
	// 建立与Redis的连接，用于缓存和会话管理
	if err := redis.Init(setting.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.Close() // 程序退出时关闭Redis连接

	// ==================== 第五步：初始化雪花算法 ====================
	// 初始化雪花算法，用于生成全局唯一的ID（如用户ID、帖子ID等）
	if err := snowflake.Init(setting.Conf.StartTime, setting.Conf.MachineID); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		return
	}

	// ==================== 第六步：初始化验证器翻译器 ====================
	// 初始化Gin框架内置验证器的中文翻译器，用于错误信息本地化
	if err := controller.InitTrans("zh"); err != nil {
		fmt.Printf("init validator trans failed, err:%v\n", err)
		return
	}

	// ==================== 第七步：设置路由并启动服务器 ====================
	// 根据运行模式（开发/生产）设置路由规则
	r := router.SetupRouter(setting.Conf.Mode)

	// 启动HTTP服务器，监听指定端口
	err := r.Run(fmt.Sprintf(":%d", setting.Conf.Port))
	if err != nil {
		fmt.Printf("run server failed, err:%v\n", err)
		return
	}
}
