/*
	基于 bt 协议的资源共享客户端服务
*/

package main

import (
	"flag"
	"log"
	"os"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/node-serv"
	"github.com/blueskyz/uvdt/node-serv/setting"
)

// 解析命令行参数配置
func ParseCmd() error {
	// http 管理界面
	httpServ := flag.String("httpserv",
		"0.0.0.0:80",
		"http server's ip and port")

	// bt 客户端之间访问接口
	btServ := flag.String("btserv",
		"0.0.0.0:8088",
		"bt server's ip and port")

	// tracker 服务器的地址
	trackerServ := flag.String("trackerserv",
		"0.0.0.0:30081",
		"tracker server's ip and port")
	logFile := flag.String("logfile",
		"/var/log/uvdt-node.log",
		"log file")

	flag.Parse()

	// 打印服务参数
	log.Printf("http server ip port: %s", *httpServ)
	log.Printf("bt server ip port: %s", *btServ)
	log.Printf("tracker server ip port: %s", *trackerServ)
	log.Printf("log file: %s", *logFile)

	// 创建配置对象
	AppSetting := &setting.AppSetting
	AppSetting.SetLogFile(*logFile)
	err := AppSetting.SetHttpServ(*httpServ)
	if err == nil {
		err = AppSetting.SetBtServ(*btServ)
	}
	if err == nil {
		err = AppSetting.SetTraceServ(*trackerServ)
	}

	return err
}

// 初始化设置
func init() {
	// 解析命令行参数
	err := ParseCmd()
	if err != nil {
		log.Printf("Error: %s", err)
		flag.Usage()
		os.Exit(-1)
	}

	// 设置日志
	logger.LoggerInit(setting.AppSetting.GetLogFile())
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()
}

func main() {
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	log.Info("node server start")

	// 启动管理服务器
	go nodeserv.BtHttpServ()

	// 启动管理服务器
	nodeserv.HttpServ()
}
