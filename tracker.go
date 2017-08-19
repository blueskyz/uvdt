/*
	tracker 服务
*/

package main

import (
	"log"
	"os"

	"github.com/blueskyz/uvdt/tracker/setting"

	"flag"
)

// 解析命令行参数配置
func ParseCmd() error {
	clusterList := flag.String("clusterip",
		"",
		"comma-separated ip list of tracker servers, ie 192.168.2.1:33000")
	btServ := flag.String("btserv",
		"0.0.0.0:80",
		"bt server ip and port")
	trackerServ := flag.String("tracserv",
		"0.0.0.0:33000",
		"tracker server ip and port for cluster")

	flag.Parse()

	// 打印服务参数
	log.Printf("cluster ip list: %s", *clusterList)
	log.Printf("bt server ip port: %s", *btServ)
	log.Printf("tracker server ip port: %s", *trackerServ)

	// 创建配置对象
	setting = setting.Setting{}
	setting.SetClusterList(clusterList)

	return nil
}

// 初始化设置
func init() {
	// 设置日志
	log.SetOutput(os.Stdout)

	// 解析命令行参数
	ParseCmd()
}

func Parse() {
}

func main() {
	log.Printf("hello world %s", "serv")
}
