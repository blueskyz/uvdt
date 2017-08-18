// tracker 服务

package main

import (
	"log"
	"os"

	"flag"
	_ "net/http"
)

func init() {
	// 设置日志
	log.SetOutput(os.Stdout)

	ipList := flag.String("serviplist", "", "tracker servers ip list, separate by comma")
	btServ := flag.String("btserv", "0.0.0.0:80", "bt server port")
	trackerServ := flag.String("tracserv", "0.0.0.0:33000", "tracker server listen for cluster")

	flag.Parse()

	// 打印服务参数
	log.Printf("ip list: %s", *ipList)
	log.Printf("bt server ip port: %s", *btServ)
	log.Printf("tracker server listen ip port: %s", *trackerServ)
}

func Parse() {
}

func main() {
	log.Printf("hello world %s", "serv")
}
