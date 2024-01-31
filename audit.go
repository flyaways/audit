package audit

import (
	"io"
	"log"
	"os"
	"syscall"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Access  Logger `toml:"access" yaml:"access" json:"access"`
	Journal Logger `toml:"journal" yaml:"journal" json:"journal"`
	Escape  string `toml:"escape" yaml:"escape" json:"escape"`
}

type Logger struct {
	Filename   string        `toml:"filename" yaml:"filename" json:"filename"`
	Level      zapcore.Level `toml:"level" yaml:"level" json:"level"`
	Rotate     string        `toml:"rotate" yaml:"rotate" json:"rotate"`
	MaxSize    int           `toml:"maxsize" yaml:"maxsize" json:"max_size"`
	MaxAge     int           `toml:"maxage" yaml:"maxage" json:"max_age"`
	MaxBackups int           `toml:"maxbackups" yaml:"maxbackups" json:"max_backups"`
	LocalTime  bool          `toml:"localtime" yaml:"localtime" json:"local_time"`
	Compress   bool          `toml:"compress" yaml:"compress" json:"compress"`
}

var (
	accessLogger  *zap.Logger //仅记录gin框架日志
	journalLogger *zap.Logger //全局默认

	accessWriter  *lumberjack.Logger //周期性日志滚动需要
	journalWriter *lumberjack.Logger //周期性日志滚动需要
)

func initAccess(writer *lumberjack.Logger, level zapcore.LevelEnabler) {
	accessWriter = writer
	accessLogger = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(
				zapcore.EncoderConfig{
					TimeKey:        "time",
					LevelKey:       "level",
					NameKey:        "logger",
					CallerKey:      "caller",
					FunctionKey:    zapcore.OmitKey,
					MessageKey:     "msg",
					StacktraceKey:  "stacktrace",
					LineEnding:     zapcore.DefaultLineEnding,
					EncodeLevel:    zapcore.LowercaseLevelEncoder,
					EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
					EncodeDuration: zapcore.SecondsDurationEncoder,
					EncodeCaller:   zapcore.ShortCallerEncoder,
				}),
			zapcore.NewMultiWriteSyncer( // write to mutil dest
				zapcore.AddSync(writer),
			),
			level),
		zap.AddCaller())
	return
}

func initJournal(writer *lumberjack.Logger, level zapcore.LevelEnabler) {
	journalWriter = writer
	journalLogger = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(
				zapcore.EncoderConfig{
					TimeKey:        "time",
					LevelKey:       "level",
					NameKey:        "logger",
					CallerKey:      "caller",
					FunctionKey:    zapcore.OmitKey,
					MessageKey:     "msg",
					StacktraceKey:  "stacktrace",
					LineEnding:     zapcore.DefaultLineEnding,
					EncodeLevel:    zapcore.LowercaseLevelEncoder,
					EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
					EncodeDuration: zapcore.SecondsDurationEncoder,
					EncodeCaller:   zapcore.ShortCallerEncoder,
				}),
			zapcore.NewMultiWriteSyncer( // write to mutil dest
				zapcore.AddSync(writer),
			),
			level),
		zap.AddCaller())

	zap.ReplaceGlobals(journalLogger) //注册为全局默认

	return
}

func AccessWriter() io.Writer {
	return accessWriter
}

func JournalWriter() io.Writer {
	return journalWriter
}

func AccessLogger() *zap.Logger {
	return accessLogger
}

func JournalLogger() *zap.Logger {
	return journalLogger
}

func initEscape(escape string) {
	file, err := os.OpenFile(escape,
		os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0644)
	if err != nil {
		log.Fatalf("initEscape :%+v", err)
	}

	syscall.Dup2(int(file.Fd()), syscall.Stdout)
	syscall.Dup2(int(file.Fd()), syscall.Stderr)
}

// Startup and init logger
func Startup(cfg *Config) {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Errorf("%+v", err)
		}
	}()

	//1 崩溃和逃逸日志
	initEscape(cfg.Escape)

	//2 初始化访问日志
	initAccess(
		&lumberjack.Logger{
			Filename:   cfg.Access.Filename,
			MaxSize:    cfg.Access.MaxSize,
			MaxAge:     cfg.Access.MaxAge,
			MaxBackups: cfg.Access.MaxBackups,
			Compress:   cfg.Access.Compress,
			LocalTime:  cfg.Access.LocalTime,
		}, cfg.Access.Level)

	//3 初始化关键日志
	initJournal(
		&lumberjack.Logger{
			Filename:   cfg.Journal.Filename,
			MaxSize:    cfg.Journal.MaxSize,
			MaxAge:     cfg.Journal.MaxAge,
			MaxBackups: cfg.Journal.MaxBackups,
			Compress:   cfg.Journal.Compress,
			LocalTime:  cfg.Journal.LocalTime,
		}, cfg.Journal.Level)

	//4 启动日志滚动
	rotate(cfg)
}

// rotate time to rotate
func rotate(cfg *Config) {
	//启动立刻滚动日志
	accessWriter.Rotate()
	journalWriter.Rotate()

	c := cron.New()
	c.AddFunc(cfg.Access.Rotate, func() {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("access log rotate: %+v", err)
			}
		}()

		log.Printf("access log rotate: %s", cfg.Access.Rotate)
		accessWriter.Rotate()
	})

	c.AddFunc(cfg.Journal.Rotate, func() {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("journal log rotate: %+v", err)
			}
		}()

		log.Printf("journal log rotate: %s", cfg.Journal.Rotate)
		journalWriter.Rotate()
	})

	c.Start()
}

// Sync to file
func Sync() {
	if accessLogger != nil {
		accessLogger.Sync()
	}

	if journalLogger != nil {
		journalLogger.Sync()
	}
}
