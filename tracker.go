/*
	tracker 服务
*/

package main

import (
	"flag"
	"log"
	"os"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/tracker/setting"
)

var app_setting = setting.Setting{}

// 解析命令行参数配置
func ParseCmd() error {
	clusterList := flag.String("clusterip",
		"",
		"comma-separated ip list of tracker servers, ie 192.168.2.1:33000")
	btServ := flag.String("btserv",
		"0.0.0.0:80",
		"bt server ip and port")
	trackerServ := flag.String("traceserv",
		"0.0.0.0:33000",
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
	app_setting.SetLogFile(*logFile)
	err := app_setting.SetClusterList(*clusterList)
	if err == nil {
		err = app_setting.SetBtServ(*btServ)
	}
	if err == nil {
		err = app_setting.SetTraceServ(*trackerServ)
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
	logger.LoggerInit(app_setting.GetLogFile())

}

func Parse() {
}

func main() {
	log.Printf("hello world %s", "serv")
}
