package utils

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Level 日志级别。建议从服务配置读取。
var LogConf = struct {
	Dir     string `yaml:"dir"`
	Name    string `yaml:"name"`
	Level   string `yaml:"level"`
	MaxSize int    `yaml:"max_size"`
}{
	Dir:     "./logs",
	Name:    "protocolgo.log",
	Level:   "trace",
	MaxSize: 100,
}

// Init logrus logger.
func InitLogger(LogLevel string) error {
	// 设置日志格式。
	logrus.SetFormatter(&CustomTextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		ForceColors:     false,
		ColorTrace:      color.New(color.FgWhite),
		ColorDebug:      color.New(color.FgCyan),
		ColorInfo:       color.New(color.FgGreen),
		ColorWarning:    color.New(color.FgYellow),
		ColorError:      color.New(color.FgRed),
		ColorCritical:   color.New(color.BgRed, color.FgWhite),
	})
	switch LogLevel {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	}
	logrus.SetReportCaller(true) // 打印文件、行号和主调函数。

	// 实现日志滚动。
	// Refer to https://www.cnblogs.com/jssyjam/p/11845475.html.
	logger := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%v/%v", LogConf.Dir, LogConf.Name), // 日志输出文件路径。
		MaxSize:    LogConf.MaxSize,                                 // 日志文件最大 size(MB)，缺省 100MB。
		MaxBackups: 100,                                             // 最大过期日志保留的个数。
		MaxAge:     30,                                              // 保留过期文件的最大时间间隔，单位是天。
		LocalTime:  true,                                            // 是否使用本地时间来命名备份的日志。
	}
	writers := []io.Writer{
		logger,
		os.Stdout}
	//同时写文件和屏幕
	fileAndStdoutWriter := io.MultiWriter(writers...)
	logrus.SetOutput(fileAndStdoutWriter)
	logrus.WithField("LogLevel", LogLevel).Info("InitLogger done.")
	return nil
}

// 自定义格式化器，继承自 logrus.TextFormatter
type CustomTextFormatter struct {
	logrus.TextFormatter
	TimestampFormat string
	ForceColors     bool
	ColorTrace      *color.Color
	ColorDebug      *color.Color
	ColorInfo       *color.Color
	ColorWarning    *color.Color
	ColorError      *color.Color
	ColorCritical   *color.Color
}

// 格式化方法，用于将日志条目格式化为字节数组
func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if f.ForceColors {
		switch entry.Level {
		case logrus.TraceLevel:
			f.ColorTrace.Println(entry.Message) // 使用蓝色打印信息日志
		case logrus.DebugLevel:
			f.ColorDebug.Println(entry.Message) // 使用蓝色打印信息日志
		case logrus.InfoLevel:
			f.ColorInfo.Println(entry.Message) // 使用蓝色打印信息日志
		case logrus.WarnLevel:
			f.ColorWarning.Println(entry.Message) // 使用黄色打印警告日志
		case logrus.ErrorLevel:
			f.ColorError.Println(entry.Message) // 使用红色打印错误日志
		case logrus.FatalLevel, logrus.PanicLevel:
			f.ColorCritical.Println(entry.Message) // 使用带有红色背景和白色文本的样式打印严重日志
		default:
			f.PrintColored(entry)
		}
		return nil, nil
	} else {
		return f.TextFormatter.Format(entry)
	}
}

// 自定义方法，用于将日志条目以带颜色的方式打印出来
func (f *CustomTextFormatter) PrintColored(entry *logrus.Entry) {
	levelColor := color.New(color.FgCyan, color.Bold)             // 定义蓝色和粗体样式
	levelText := levelColor.Sprintf("%-6s", entry.Level.String()) // 格式化日志级别文本

	msg := levelText + " " + entry.Message
	if entry.HasCaller() {
		msg += " (" + entry.Caller.File + ":" + strconv.Itoa(entry.Caller.Line) + ")" // 添加调用者信息
	}

	fmt.Fprintln(color.Output, msg) // 使用有颜色的方式打印消息到终端
}
