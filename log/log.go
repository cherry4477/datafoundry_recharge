package log

import (
	"flag"
	"github.com/astaxie/beego/logs"
	"os"
	"strings"
)

const (
	CHANNELLEN = 3000
)

var (
	loglevel = os.Getenv("LOG_LEVEL")
	logger   *logs.BeeLogger
)

func init() {
	flag.Parse()
	logger = logs.NewLogger(CHANNELLEN)

	err := logger.SetLogger("console", "")
	if err != nil {
		logger.Error("set logger err:", err)
		return
	}

	//显示文件名和行号
	logger.EnableFuncCallDepth(true)

	//判断是不是以 DEBUG 模式启动
	if strings.ToUpper(loglevel) == "DEBUG" {
		logger.Info("mode is info...")
		logger.SetLevel(logs.LevelDebug)
	} else {
		logger.Info("mode is debug...")
		logger.SetLevel(logs.LevelInfo)
	}
}

func GetLogger() *logs.BeeLogger {
	return logger
}
