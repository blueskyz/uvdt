/*
	tracker bt http server 服务，提供 node peer 访问
*/
package tracker

import (
	"fmt"
	"net/http"
	"strconv"
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
	btHttpServMux.HandleFunc("/node", btNodeHandler)
	btHttpServMux.HandleFunc("/", btTorrentHandler)

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

func btNodeHandler(w http.ResponseWriter, r *http.Request) {

	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	log.Info(r.RequestURI)
	// 解析 bt 请求参数
	values := r.URL.Query()
	if len(values) == 0 {
		CreateErrResp(w, &log, "Arguments is empty")
		return
	}

	infoHash := values.Get("info_hash")
	if !CheckHexdigest(infoHash, 32) {
		CreateErrResp(w, &log, "infoHash err")
		return
	}

	compact := values.Get("compact")

	// 获取 peer id, ip, port 信息
	peerId := values.Get("peer_id")
	if !CheckHexdigest(peerId, 32) {
		CreateErrResp(w, &log, "peer id's length is not 20")
		return
	}

	ip := strings.Split(r.RemoteAddr, ":")[0]
	port := values.Get("port")
	port_int, err := strconv.Atoi(port)
	if err != nil || port_int < 0 || port_int > 65535 {
		CreateErrResp(w, &log, fmt.Sprintf("Port is err: %s", port))
		return
	}
	log.Info(fmt.Sprintf("infoHash: %s, compact: %s, peerId: %s, ip: %s, port: %s",
		infoHash, compact, peerId, ip, port))

	// 检查保存 node 信息

	// 获取 peer list
	info := Torrent{
		infoHash: infoHash,
		name:     "",
		peer:     fmt.Sprintf("%s:%s:%s", peerId, ip, port),
	}
	peers, err := info.GetPeers(infoHash)
	if err != nil {
		CreateErrResp(w, &log, fmt.Sprintf("Get peers err: %s", err))
		return
	}

	// 创建 response
	btResp := map[string]interface{}{
		"info_hash": infoHash,
		"peers":     peers,
		"interval":  30,
	}

	CreateSuccResp(w, &log, "succ", btResp)
}

func btTorrentHandler(w http.ResponseWriter, r *http.Request) {

	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	// fixme: 必须登陆

	values := r.URL.Query()
	if len(values) == 0 {
		CreateErrResp(w, &log, "Arguments is empty")
		return
	}
	infoHash := values.Get("infohash")
	if !CheckHexdigest(infoHash, 32) {
		hashLen := len(infoHash)
		if hashLen > 32 {
			infoHash = infoHash[:33]
		}
		CreateErrResp(w, &log, fmt.Sprintf("infoHash parameter err, infohash=%s, len=%d",
			hashLen,
			infoHash))
		return
	}

	// 获取 peer id, ip, port 信息
	peerId := values.Get("peer_id")
	if !CheckHexdigest(peerId, 32) {
		CreateErrResp(w, &log, "peer id's length is not 20")
		return
	}

	ip := strings.Split(r.RemoteAddr, ":")[0]
	port := values.Get("port")
	port_int, err := strconv.Atoi(port)
	if err != nil || port_int < 0 || port_int > 65535 {
		CreateErrResp(w, &log, fmt.Sprintf("Port is err: %s", port))
		return
	}
	log.Info(fmt.Sprintf("infoHash: %s, peerId: %s, ip: %s, port: %s",
		infoHash, peerId, ip, port))

	if r.Method == "GET" {
		// 获取 torrent file
		log.Info(fmt.Sprintf("GET: infohash=%s", infoHash))

		/*
			torrent := Torrent{
				infoHash: infoHash,
				name:     "",
				peer:     fmt.Sprintf("%s:%s:%s", peerId, ip, port),
			}
			msg, err := info.AddTorrent(torrent)
			if err != nil {
				CreateErrResp(w, &log, fmt.Sprintf("Get torrent err: %s", err))
				return
			}
			w.Write()
		*/
	} else if r.Method == "POST" {
		// 上传 torrent file
		log.Info(fmt.Sprintf("POST: infohash=%s", infoHash))
		torrent := values.Get("torrent")
		if len(torrent) >= (1024 << 12) {
			CreateErrResp(w, &log, fmt.Sprintf("infohash=%s, torrent len=%d",
				infoHash,
				len(torrent)))
			return
		}

		info := Torrent{
			infoHash: infoHash,
			name:     "",
			peer:     fmt.Sprintf("%s:%s:%s", peerId, ip, port),
		}
		torrent, err := info.GetTorrent(infoHash)
		if err != nil {
			CreateErrResp(w, &log, fmt.Sprintf("Get torrent err: %s", err))
			return
		}
	}

	// 创建 response
	btResp := map[string]interface{}{
		"info_hash": infoHash,
	}
	CreateSuccResp(w, &log, "succ", btResp)
}
