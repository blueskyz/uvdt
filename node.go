/*
	基于 bt 协议的资源共享客户端服务
*/

package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/node-serv"
	"github.com/blueskyz/uvdt/node-serv/setting"
)

// 解析命令行参数配置
func ParseCmd() error {
	// 设置 root 目录，所有共享的文件必须在这个目录里，
	// 并且提供给 node-serv 服务使用
	rootPath := flag.String("rootpath",
		"",
		"root path for all resource file")

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
	if len(*rootPath) == 0 {
		return errors.New("root path is empty")
	} else {
		AppSetting.SetRootPath(*rootPath)
	}

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
	logAgent := logger.NewAgent()
	defer logAgent.EndLog()

	logAgent.Info("node server start")

	// 1. 创建下载和分享的文件对象
	// 2. 启动下载服务
	filesMgr, err := nodeserv.CreateFilesMgr()
	if err != nil {
		log.Printf("Err: %s", err.Error())
		flag.Usage()
		os.Exit(-1)
	}

	// 启动资源分享服务器
	go nodeserv.BtHttpServ(filesMgr)

	// 启动管理服务器
	err = nodeserv.HttpServ(filesMgr)
	if err != nil {
		log.Printf("Err: %s", err.Error())
		flag.Usage()
		os.Exit(-1)
	}
}
