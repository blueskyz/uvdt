/*
	tracker 服务
*/

package main

import (
	"flag"
	"log"
	"os"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/tracker"
	"github.com/blueskyz/uvdt/tracker/setting"
)

// 解析命令行参数配置
func ParseCmd() error {
	clusterList := flag.String("clusterip",
		"",
		"comma-separated ip list of tracker servers, ie 192.168.2.1:30081")
	btServ := flag.String("btserv",
		"0.0.0.0:80",
		"bt server ip and port")
	trackerServ := flag.String("trackerserv",
		"0.0.0.0:30081",
		"tracker server ip and port for cluster")
	logFile := flag.String("logfile",
		"/var/log/uvdt-trace.log",
		"log file")

	flag.Parse()

	// 打印服务参数
	log.Printf("cluster ip list: %s", *clusterList)
	log.Printf("bt server ip port: %s", *btServ)
	log.Printf("tracker server ip port: %s", *trackerServ)
	log.Printf("log file: %s", *logFile)

	// 创建配置对象
	AppSetting := &setting.AppSetting
	AppSetting.SetLogFile(*logFile)
	err := AppSetting.SetClusterList(*clusterList)
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

	// 设置数据库
	err = tracker.InitDB("127.0.0.1:3306", "root", "1")
	if err != nil {
		log.Err(err.Error())
	}

	// 设置数缓存
	tracker.InitRedis("127.0.0.1:6379", "11")
}

func main() {
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	log.Info("hello world serv")

	// 启动管理服务器
	go tracker.TrackerHttpServ()

	// 启动 bt http 服务
	tracker.BtHttpServ()
}
