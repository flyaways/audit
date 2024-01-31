package main

import (
	"fmt"
	"log"

	"github.com/flyaways/audit"
	"go.uber.org/zap"
)

func main() {
	// init test config
	cfg := initConfig()

	// init audit log
	audit.Startup(cfg)
	defer audit.Sync()

	// //TODO: setting to replace
	// gin.DefaultWriter = audit.AccessWriter() //access.log
	// // gin.DefaultErrorWriter = audit.JournalWriter() //default to escape.log
	// log.SetOutput(audit.JournalWriter()) //journal.log

	// //TODO:
	// r := gin.Default()
	// r.Any("/*path", func(c *gin.Context) {
	// 	param := c.Param("path")
	// 	log.Println(param)

	// 	if param == "/panic" {
	// 		panic("PING")
	// 	}

	// 	c.JSON(http.StatusOK, gin.H{})
	// })
	// r.Run()

	journal := zap.L()
	journal.Info("say", zap.String("hello", "world")) // journal.log

	// do what you wana to do ......
	fmt.Println("say to fmt hello") //escape.log
	log.Println("say to access")    //access.log
	panic("something")              //escape.log
}

func initConfig() *audit.Config {
	return &audit.Config{
		Access: audit.Logger{
			Filename:   "access.log", //访问日志
			Level:      -1,           //DebugLevel:-1,InfoLevel:0,WarnLevel:1,ErrorLevel:2,DPanicLevel:3,PanicLevel:4,FatalLevel:5
			Rotate:     "@every 15h", //"@midnight"
			MaxSize:    1024,         //m
			MaxAge:     7,            //days
			MaxBackups: 7,            //days
			LocalTime:  true,
			Compress:   false,
		},
		Journal: audit.Logger{
			Filename:   "journal.log", //关键业务逻辑日志
			Level:      -1,            //DebugLevel:-1,InfoLevel:0,WarnLevel:1,ErrorLevel:2,DPanicLevel:3,PanicLevel:4,FatalLevel:5
			Rotate:     "@every 15h",  //"@midnight"
			MaxSize:    1024,          //m
			MaxAge:     7,             //days
			MaxBackups: 7,             //days
			LocalTime:  true,
			Compress:   false,
		},
		Escape: "escape.log", //崩溃和逃逸日志",
	}
}
