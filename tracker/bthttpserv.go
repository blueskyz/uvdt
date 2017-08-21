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
	defer log.EndLog()

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

	infoHash := values.Get("info_hash")
	if len(infoHash) == 0 {
		log.Err("infoHash is empty")
	}

	compact := values.Get("compact")

	peerId := values.Get("peer_id")
	if len(peerId) == 0 {
		log.Err("peerId is empty")
	}

	ip := strings.Split(r.RemoteAddr, ":")[0]
	port := values.Get("port")
	if len(port) == 0 {
		log.Err("port is empty")
	}
	log.Info(fmt.Sprintf("infoHash: %s, compact: %s, peerId: %s, ip: %s, port: %s",
		infoHash, compact, peerId, ip, port))

	// 检查保存 node 信息

	// 获取 peer list
	info := InfoHash{
		infoHash: infoHash,
		name:     "",
		peer:     fmt.Sprintf("%s:%s:%s", peerId, ip, port),
	}
	peers, err := info.GetInfoHash(infoHash)
	if err != nil {
		log.Err(fmt.Sprintf("Get info hash err: %s", err))
	}
	fmt.Printf("peers: %v\n", peers)
}
