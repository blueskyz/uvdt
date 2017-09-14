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
	"github.com/blueskyz/uvdt/node-tool"
	"github.com/blueskyz/uvdt/node-tool/setting"
)

// 解析命令行参数配置
func ParseCmd() error {
	// 设置 root 目录，所有共享的文件必须在这个目录里，
	// 并且提供给 node-serv 服务使用
	rootPath := flag.String("rootpath",
		"",
		"root path for all resource file")

	// 添加共享文件资源目录，遍历目录下的文件，
	// 创建 infohash，并且提交到 tracker 服务器
	resPath := flag.String("respath",
		"",
		"add a shared resource Path")

	// 添加共享文件资源，创建 infohash，并且提交到 tracker 服务器
	resFile := flag.String("resfile",
		"",
		"add a shared resource file")

	// tracker 服务器的地址
	trackerServ := flag.String("trackerserv",
		"0.0.0.0:30081",
		"tracker server's ip and port")
	logFile := flag.String("logfile",
		"/var/log/uvdt-node-tool.log",
		"log file")

	flag.Parse()

	// 打印服务参数
	log.Printf("root path: %s", *rootPath)
	log.Printf("add a shared resource path: %s", *resPath)
	log.Printf("add a shared resource file: %s", *resFile)
	log.Printf("tracker server ip port: %s", *trackerServ)
	log.Printf("log file: %s", *logFile)

	// 创建配置对象
	AppSetting := &setting.AppSetting
	AppSetting.SetLogFile(*logFile)

	if len(*rootPath) == 0 {
		return errors.New("root path is empty")
	} else {
		AppSetting.SetRootPath(*rootPath)
	}

	fileInfo, err := os.Stat(*rootPath)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return errors.New("root path is not directory")
	}

	if len(*resPath) == 0 {
		log.Printf("resource path is empty")
	} else {
		AppSetting.SetResPath(*resPath)
	}
	if len(*resFile) == 0 {
		log.Printf("resource file is empty")
	} else {
		AppSetting.SetResFile(*resFile)
	}
	if len(*resPath) == 0 && len(*resFile) == 0 {
		return errors.New("both resource path and resource file are empty")
	}
	err = AppSetting.SetTraceServ(*trackerServ)

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
}

func main() {
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	log.Info("hello world serv")

	creatorTorrent := nodetool.CreatorTorrent{}
	_, err := creatorTorrent.ScanPath()
	if err != nil {
		log.Err(err.Error())
	}
}
