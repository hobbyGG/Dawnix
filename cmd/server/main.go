package main

import (
	"os"
	"strings"

	"github.com/hobbyGG/Dawnix/internal/app"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func main() {
	// zap日志库初始化
	logger := NewLogger()
	defer logger.Sync()

	// 创建并运行应用程序
	app, err := app.NewAppManual(logger)
	if err != nil {
		panic(err)
	}
	defer func() {
		if app.Cleanup != nil {
			app.Cleanup()
		}
	}()

	app.Run()
}

func NewLogger() *zap.Logger {
	// 1. 配置基础的 Encoder Config
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	// 这里我们把 Level 的显示设为大写，但不需要它自带的颜色了，因为我们要自己整行加颜色
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	// 也可以自定义时间格式
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")

	// 2. 创建一个标准的 ConsoleEncoder
	baseEncoder := zapcore.NewConsoleEncoder(encoderCfg)

	// 3. 【关键】使用自定义的 ColoredEncoder 包装标准 Encoder
	coloredEncoder := &ColoredEncoder{Encoder: baseEncoder}

	// 4. 创建 Core
	core := zapcore.NewCore(coloredEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel)
	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)
	return logger
}

// ==========================================
// 自定义 Encoder 装饰器
// ==========================================

// ANSI 颜色代码
const (
	ColorRed    = "\x1b[31m"
	ColorGreen  = "\x1b[32m"
	ColorYellow = "\x1b[33m" // 终端里的 Orange 通常就是 Yellow
	ColorReset  = "\x1b[0m"
)

// ColoredEncoder 结构体嵌入了 zapcore.Encoder 接口
// 这样我们只需要重写 EncodeEntry 方法，其他方法直接使用原本的 ConsoleEncoder 实现
type ColoredEncoder struct {
	zapcore.Encoder
}

// EncodeEntry 是日志输出的核心方法
func (e *ColoredEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// 1. 先让原本的 ConsoleEncoder 把日志格式化好（带时间、Caller、Message 等）
	buf, err := e.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	// 2. 根据日志级别选择颜色
	var color string
	switch entry.Level {
	case zapcore.InfoLevel:
		color = ColorGreen
	case zapcore.WarnLevel:
		color = ColorYellow
	case zapcore.ErrorLevel:
		color = ColorRed
	default:
		// 其他级别（如 Debug）不改变颜色，或者你可以自己加
		return buf, nil
	}

	// 3. 构造新的 buffer 对整行进行包装
	// Zap 的 EncodeEntry 通常会在末尾带一个换行符，我们需要处理一下
	content := buf.String()

	// 如果内容以换行符结尾，去掉它，以便我们将 Reset 代码放在换行符之前
	// 这样可以防止背景色溢出到下一行
	content = strings.TrimSuffix(content, "\n")

	// 重新写入：颜色 + 原内容 + 重置颜色 + 换行
	newBuf := buffer.NewPool().Get()
	newBuf.AppendString(color)
	newBuf.AppendString(content)
	newBuf.AppendString(ColorReset)
	newBuf.AppendString("\n")

	// 释放原本的 buffer
	buf.Free()

	return newBuf, nil
}

// Clone 方法必须重写，因为 logger.With() 会调用它
func (e *ColoredEncoder) Clone() zapcore.Encoder {
	return &ColoredEncoder{
		Encoder: e.Encoder.Clone(),
	}
}
