/*
	tracker bt http server 服务，提供 node peer 访问
*/
package tracker

import (
	"fmt"
	"io/ioutil"
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
	btHttpServMux.HandleFunc("/", btHelloHandler)
	btHttpServMux.HandleFunc("/node", btNodeHandler)
	btHttpServMux.HandleFunc("/torrent", btTorrentHandler)

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
		CreateErrResp(w, &log, "peer id's length is not 32")
		return
	}

	ip := strings.Split(r.RemoteAddr, ":")[0]
	port := values.Get("port")
	port_int, err := strconv.Atoi(port)
	if err != nil || port_int < 0 || port_int > 65535 {
		CreateErrResp(w, &log, fmt.Sprintf("Port[%s] is err: %s", port, err.Error()))
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
			infoHash,
			hashLen))
		return
	}

	// 获取 peer id, ip, port 信息
	peerId := values.Get("peer_id")
	if !CheckHexdigest(peerId, 32) {
		CreateErrResp(w, &log, "peer id's length is not 32")
		return
	}

	// 获取 ip
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if !CheckIP(ip) {
		ip = "127.0.0.1"
	}

	// 获取 port
	port := values.Get("port")
	port_int, err := strconv.Atoi(port)
	if err != nil || port_int < 0 || port_int > 65535 {
		CreateErrResp(w, &log, fmt.Sprintf("Port is err: %s", err.Error()))
		return
	}
	log.Info(fmt.Sprintf("infoHash: %s, peerId: %s, ip: %s, port: %s",
		infoHash, peerId, ip, port))

	// 创建 response
	btResp := map[string]interface{}{
		"info_hash": infoHash,
	}

	if r.Method == "GET" {
		// 获取 torrent file
		log.Info(fmt.Sprintf("GET: infohash=%s", infoHash))

		torrent := Torrent{
			infoHash: infoHash,
			name:     "",
			peer:     fmt.Sprintf("%s:%s:%s", peerId, ip, port),
		}
		torrentContent, err := torrent.GetTorrent(infoHash)
		if err != nil {
			CreateErrResp(w, &log, fmt.Sprintf("Get torrentContent err: %s", err))
			return
		}
		btResp["torrent_content"] = torrentContent
	} else if r.Method == "POST" {
		// 上传 torrent file
		/*
			r.ParseForm()
			postValues := r.PostForm
			log.Info(fmt.Sprintf("POST: infohash=%s", infoHash))
			torrentContent := postValues.Get("torrent")
		*/
		body, err := ioutil.ReadAll(r.Body)
		torrentContent := string(body)
		if len(torrentContent) >= (1024 << 12) {
			CreateErrResp(w, &log, fmt.Sprintf("too big, infohash=%s, torrentContent len=%d",
				infoHash,
				len(torrentContent)))
			return
		}

		if len(torrentContent) == 0 {
			CreateErrResp(w, &log, fmt.Sprintf("too short, infohash=%s, torrentContent len=%d",
				infoHash,
				len(torrentContent)))
			return
		}

		torrent := Torrent{
			infoHash: infoHash,
			name:     "",
			peer:     fmt.Sprintf("%s:%s:%s", peerId, ip, port),
		}
		log.Info(fmt.Sprintf("content=%s", torrentContent))
		content, err := ParseBtProto(infoHash, torrentContent)
		if err != nil {
			log.Err(err.Error())
			CreateErrResp(w, &log, "Parse torrent fail")
			return
		}

		log.Info(fmt.Sprintf("version=%s", content["version"].(string)))
		err = torrent.AddTorrent(infoHash, torrentContent)
		if err != nil {
			log.Err(err.Error())
			CreateErrResp(w, &log, "upload torrent to db error")
			return
		}
	}

	CreateSuccResp(w, &log, "succ", btResp)
}
