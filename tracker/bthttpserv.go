/*
	tracker bt http server 服务，提供 node peer 访问
*/
package tracker

import (
	"encoding/json"
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

// 错误输出函数
func toErrRsp(w http.ResponseWriter, log logger.LogAgent, errMsg string) {
	log.Err(errMsg)
	errResp := CreateErrResp(-1, errMsg)
	w.Write(errResp)
}

func btNodeHandler(w http.ResponseWriter, r *http.Request) {

	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	log.Info(r.RequestURI)
	// 解析 bt 请求参数
	values := r.URL.Query()
	if len(values) == 0 {
		toErrRsp(w, log, "Arguments is empty")
		return
	}

	infoHash := values.Get("info_hash")
	if !CheckHexdigest(infoHash, 32) {
		toErrRsp(w, log, "infoHash err")
		return
	}

	compact := values.Get("compact")

	peerId := values.Get("peer_id")
	if !CheckHexdigest(peerId, 20) {
		toErrRsp(w, log, "peer id's length is not 20")
		return
	}

	ip := strings.Split(r.RemoteAddr, ":")[0]
	port := values.Get("port")
	port_int, err := strconv.Atoi(port)
	if err != nil || port_int < 0 || port_int > 65535 {
		toErrRsp(w, log, fmt.Sprintf("Port is err: %s", port))
		return
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
	peers, err := info.GetPeersFromInfoHash(infoHash)
	if err != nil {
		errMsg := fmt.Sprintf("Get info hash err: %s", err)
		log.Err(errMsg)
		errResp := CreateErrResp(-1, "Get peers fail.")
		w.Write(errResp)
		return
	}

	// 创建 response
	btResp := map[string]interface{}{
		"info_hash": infoHash,
		"peers":     peers,
		"interval":  30,
	}
	rspBody, err := json.Marshal(btResp)
	w.Header().Set("Content-type", "application/json")
	if err != nil {
		errMsg := fmt.Sprintf("json serialze fail: %s", err.Error())
		log.Err(errMsg)
		errResp := CreateErrResp(-1, "json serialize fail.")
		w.Write(errResp)
	} else {
		w.Write(rspBody)
	}
}

func btTorrentHandler(w http.ResponseWriter, r *http.Request) {

	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	values := r.URL.Query()
	if len(values) == 0 {
		toErrRsp(w, log, "Arguments is empty")
		return
	}
	infoHash := values.Get("infohash")
	if !CheckHexdigest(infoHash, 32) {
		outInfoHash := [:]byte{}
		if len(infoHash) > 32 {
			outInfoHash := infoHash[:32]
		}
		toErrRsp(w, log, fmt.Sprint("infoHash parameter err, infohash=%s", out_infoHash[39]))
		return
	}

	// 获取 torrent file
	log.Info(fmt.Sprintf("%s: infohash=%s", r.Method, infoHash))
	if r.Method == "GET" {
		log.Info(fmt.Sprintf("GET: infohash=%s", infoHash))
	}

	// 上传 torrent file
	if r.Method == "POST" {
		log.Info(fmt.Sprintf("POST: infohash=%s", infoHash))
	}
}
