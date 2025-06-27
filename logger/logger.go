// Package logger 提供日志记录功能
// 基于zap日志库实现结构化日志记录，支持文件轮转和开发/生产模式
package logger

import (
	"bluebell/setting"  // 导入配置包，获取日志配置信息
	"net"               // 导入网络包，用于检测网络错误
	"net/http"          // 导入HTTP包，提供HTTP状态码
	"net/http/httputil" // 导入HTTP工具包，用于转储HTTP请求
	"os"                // 导入操作系统包，用于标准输出和系统错误检测
	"runtime/debug"     // 导入调试包，用于获取堆栈信息
	"strings"           // 导入字符串包，用于字符串处理
	"time"              // 导入时间包，用于时间计算

	"github.com/gin-gonic/gin"        // 导入Gin框架，用于HTTP中间件
	"github.com/natefinch/lumberjack" // 导入日志轮转包，实现日志文件自动轮转
	"go.uber.org/zap"                 // 导入zap日志库，提供高性能结构化日志
	"go.uber.org/zap/zapcore"         // 导入zap核心包，提供日志核心功能
)

// lg 全局日志记录器实例
// 使用zap.Logger类型，提供高性能的结构化日志记录
var lg *zap.Logger

// Init 初始化日志系统
// 根据配置信息设置日志级别、输出位置、格式等
// 参数 cfg: 日志配置信息，包含文件名、大小限制、备份数量等
// 参数 mode: 运行模式（dev/prod），影响日志输出方式
// 返回值: 错误信息，成功时返回nil
func Init(cfg *setting.LogConfig, mode string) (err error) {
	// ==================== 第一步：配置日志输出器 ====================
	// 获取日志写入器，支持文件轮转功能
	writeSyncer := getLogWriter(cfg.Filename, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)

	// ==================== 第二步：配置日志编码器 ====================
	// 获取JSON格式的日志编码器
	encoder := getEncoder()

	// ==================== 第三步：解析日志级别 ====================
	// 将字符串格式的日志级别转换为zapcore.Level类型
	var l = new(zapcore.Level)
	err = l.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return
	}

	// ==================== 第四步：创建日志核心 ====================
	var core zapcore.Core
	if mode == "dev" {
		// 开发模式：日志同时输出到文件和终端
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, writeSyncer, l),                                     // 文件输出
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel), // 终端输出
		)
	} else {
		// 生产模式：日志只输出到文件
		core = zapcore.NewCore(encoder, writeSyncer, l)
	}

	// ==================== 第五步：创建日志记录器 ====================
	// 创建日志记录器实例，添加调用者信息
	lg = zap.New(core, zap.AddCaller())

	// 替换全局日志记录器
	zap.ReplaceGlobals(lg)
	zap.L().Info("init logger success")
	return
}

// getEncoder 获取日志编码器
// 配置JSON格式的日志输出，包含时间、级别、调用者等信息
// 返回值: 配置好的JSON编码器
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder         // 使用ISO8601时间格式
	encoderConfig.TimeKey = "time"                                // 时间字段名
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder       // 大写级别编码器
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder // 秒级持续时间编码器
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder       // 短格式调用者编码器
	return zapcore.NewJSONEncoder(encoderConfig)
}

// getLogWriter 获取日志写入器
// 配置日志文件轮转功能，支持按大小、数量、时间自动轮转
// 参数 filename: 日志文件名
// 参数 maxSize: 单个日志文件最大大小（MB）
// 参数 maxBackup: 最大备份文件数量
// 参数 maxAge: 日志文件最大保存天数
// 返回值: 配置好的日志写入器
func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,  // 日志文件路径
		MaxSize:    maxSize,   // 单个文件最大大小（MB）
		MaxBackups: maxBackup, // 最大备份文件数量
		MaxAge:     maxAge,    // 日志文件最大保存天数
	}
	return zapcore.AddSync(lumberJackLogger)
}

// GinLogger 接收gin框架默认的日志中间件
// 记录HTTP请求的详细信息，包括状态码、方法、路径、耗时等
// 返回值: Gin中间件函数
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算请求耗时
		cost := time.Since(start)

		// 记录请求日志
		lg.Info(path,
			zap.Int("status", c.Writer.Status()),                                 // HTTP状态码
			zap.String("method", c.Request.Method),                               // HTTP方法
			zap.String("path", path),                                             // 请求路径
			zap.String("query", query),                                           // 查询参数
			zap.String("ip", c.ClientIP()),                                       // 客户端IP
			zap.String("user-agent", c.Request.UserAgent()),                      // 用户代理
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()), // 私有错误
			zap.Duration("cost", cost),                                           // 请求耗时
		)
	}
}

// GinRecovery 恢复项目可能出现的panic，并使用zap记录相关日志
// 防止程序因panic而崩溃，记录详细的错误信息和堆栈

func GinRecovery(stack bool) gin.HandlerFunc {
	// 返回一个Gin中间件函数，用于处理panic
	// 参数 stack: 是否记录堆栈信息, 堆栈信息就是程序执行过程中函数调用的路径图
	// 每行堆栈信息包括函数名、文件名、行号，可以快速定位到具体的代码位置; 堆栈信息是字节数组，需要转换为字符串
	// 返回值: Gin中间件函数
	return func(c *gin.Context) {
		// defer func() { ... }() 是Go语言的延迟执行语法
		// 无论函数正常结束还是发生panic，defer中的代码都会执行
		// 这就像是一个"保险"，确保panic被捕获和处理
		defer func() {
			// recover() 是Go语言的内置函数，用于捕获panic
			// 当程序发生panic时，recover()会返回panic的值
			// 如果程序正常运行，recover()返回nil
			if err := recover(); err != nil {
				// ==================== 第一步：检测网络连接错误 ====================
				// 检查是否为网络连接断开错误，这类错误不需要记录堆栈信息
				// 因为网络断开是正常现象，不是程序bug
				var brokenPipe bool // 是否为网络连接断开错误

				// 类型断言语法：err.(*net.OpError)
				// 尝试将err转换为*net.OpError类型
				// 如果转换成功，ok为true；如果失败，ok为false
				if ne, ok := err.(*net.OpError); ok {
					// 再次类型断言，检查是否是系统调用错误
					if se, ok := ne.Err.(*os.SyscallError); ok {
						// strings.ToLower() 将字符串转换为小写
						// strings.Contains() 检查字符串是否包含子串
						// 检查错误信息是否包含"broken pipe"或"connection reset by peer"
						// 这些都是网络连接断开的常见错误信息
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// ==================== 第二步：转储HTTP请求信息 ====================
				// httputil.DumpRequest() 将HTTP请求转换为字符串格式
				// 第二个参数false表示不包含请求体，只包含请求头
				// 这样可以在日志中看到请求的基本信息，但不会记录敏感数据
				httpRequest, _ := httputil.DumpRequest(c.Request, false)

				if brokenPipe {
					// 网络连接错误，记录简要信息
					// 因为连接已断开，无法向客户端发送响应
					lg.Error(c.Request.URL.Path,
						zap.Any("error", err),                      // 记录错误信息
						zap.String("request", string(httpRequest)), // 记录请求信息
					)
					// 连接已断开，无法写入状态码
					// c.Error() 将错误添加到Gin的错误列表中
					// c.Abort() 终止后续中间件和处理器执行
					c.Error(err.(error)) // 类型断言：将err转换为error类型
					c.Abort()
					return
				}

				// ==================== 第三步：记录panic错误信息 ====================
				if stack {
					// 记录完整的堆栈信息
					// debug.Stack() 获取当前的调用堆栈信息
					// 堆栈信息显示函数调用的完整路径，有助于调试
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),                      // 记录panic的值
						zap.String("request", string(httpRequest)), // 记录请求信息
						zap.String("stack", string(debug.Stack())), // 记录堆栈信息
					)
				} else {
					// 只记录错误信息，不记录堆栈
					// 生产环境可能不记录堆栈信息，避免日志文件过大
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),                      // 记录panic的值
						zap.String("request", string(httpRequest)), // 记录请求信息
					)
				}

				// 返回500内部服务器错误
				// c.AbortWithStatus() 终止请求处理并返回指定的HTTP状态码
				// http.StatusInternalServerError 是500状态码的常量
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		// c.Next() 继续执行后续的中间件和请求处理函数
		// 如果后续代码发生panic，会被上面的defer函数捕获
		c.Next()
	}
}
