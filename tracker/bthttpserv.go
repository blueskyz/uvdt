/*
	tracker bt http server 服务，提供 node peer 访问
*/
package tracker

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/tracker/setting"
)

func BtHttpServ() {
	log := logger.NewAgent()
	// 设置 bt http server 路由
	btHttpServMux := http.NewServeMux()
	btHttpServMux.HandleFunc("/hello", btHelloHandler)
	btHttpServMux.HandleFunc("/node", btHandler)

	btServ := setting.AppSetting.GetBtServ()
	log.Info(fmt.Sprintf("init %s:%d", btServ.Ip, btServ.Port))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", btServ.Ip, btServ.Port), btHttpServMux)
	if err != nil {
		log.Err("init, " + err.Error())
	}
}

func btHelloHandler(w http.ResponseWriter, r *http.Request) {
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	fmt.Fprintf(w, "hello bt http serv")
}

func btHandler(w http.ResponseWriter, r *http.Request) {
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	log.Info(r.RequestURI)
	// 解析 bt 参数
	values := r.URL.Query()
	if len(values) == 0 {
		log.Err("arguments err")
	}

	info_hash := values.Get("info_hash")
	if len(info_hash) == 0 {
		log.Err("info_hash is empty")
	}

	compact := values.Get("compact")

	peer_id := values.Get("peer_id")
	if len(peer_id) == 0 {
		log.Err("peer_id is empty")
	}

	ip := strings.Split(r.RemoteAddr, ":")[0]
	port := values.Get("port")
	if len(port) == 0 {
		log.Err("port is empty")
	}
	log.Info(fmt.Sprintf("info_hash: %s, compact: %s, peer_id: %s, ip: %s, port: %s",
		info_hash, compact, peer_id, ip, port))

	// 检查保存 node 信息

	// 获取 peer list
}
