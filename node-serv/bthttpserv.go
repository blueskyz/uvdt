/*
	bt http server 服务，提供下载服务
*/
package nodeserv

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/node-serv/setting"
	"github.com/blueskyz/uvdt/utils"
)

var btFilesMgr *FilesManager

func BtHttpServ(filesManager *FilesManager) error {
	log := logger.NewAgent()
	defer log.EndLog()

	btFilesMgr = filesManager

	// 设置  http server 路由
	HttpBtServMux := http.NewServeMux()
	HttpBtServMux.HandleFunc("/hello", httpBtHelloHandler)

	// 通过创建分享任务
	HttpBtServMux.HandleFunc("/api/share/resource", httpShareResourceHandler)

	// HttpBtServMux.HandleFunc("/test", httpBtTestHandler)

	// 提供文件分片下载服务
	HttpBtServMux.HandleFunc("/api/download", httpBtDownloadHandler)

	httpBtServ := setting.AppSetting.GetBtServ()
	log.Info(fmt.Sprintf("%s:%d", httpBtServ.Ip, httpBtServ.Port))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d",
		httpBtServ.Ip,
		httpBtServ.Port),
		HttpBtServMux)
	if err != nil {
		log.Err(err.Error())
	}
	return err
}

func httpBtHelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "bt http serv hello")
}

func httpBtTestHandler(w http.ResponseWriter, r *http.Request) {
	showDownLoadStat := fmt.Sprintf("bt test http btFilesMgr maxFileNum=%d, currentNum=%d, ",
		btFilesMgr.GetMaxFileNum(),
		btFilesMgr.GetCurrentFileNum())
	w.Write([]byte(showDownLoadStat))
}

/*
 * 下载资源
 */
func httpBtDownloadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "http serv api download")
}

/*
 * 共享资源
 */
func httpShareResourceHandler(w http.ResponseWriter, r *http.Request) {
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	log.Info(r.RequestURI)
	// 解析 bt 请求参数
	values := r.URL.Query()
	if len(values) == 0 {
		utils.CreateErrResp(w, &log, "Arguments is empty")
		return
	}

	infoHashName := values.Get("info_hash_name")
	sharePath := path.Join(setting.AppSetting.GetRootPath(), "share", ".torrents")

	// 1. 读取 torrent file
	torrent_file := path.Join(sharePath, infoHashName)
	log.Info(torrent_file)
	if _, err := os.Stat(torrent_file); os.IsNotExist(err) {
		log.Err(fmt.Sprintf("The share torrent file is not exist, %s", torrent_file))
		utils.CreateErrResp(w, &log, "The share torrent file is not exist")
		return
	} else if os.IsExist(err) {
		log.Err(fmt.Sprintf("The share torrent file error, %s", torrent_file))
		utils.CreateErrResp(w, &log, "Open share torrent file error")
		return
	}

	f, err := os.OpenFile(torrent_file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Err(fmt.Sprintf("%s: %s", f, err.Error()))
		utils.CreateErrResp(w, &log, "Open share torrent file error")
		return
	}
	defer f.Close()

	torFile := make([]byte, 1024<<10)
	count, err := f.Read(torFile)
	if err != nil && err != io.EOF {
		log.Err(fmt.Sprintf("Read tor data fail, %s", torrent_file))
		utils.CreateErrResp(w, &log, "Read share torrent file error")
		return
	}
	torFile = torFile[:count]

	// 2. 从 share 目录找到共享的文件
	//	  创建本地共享文件
	_, infohash, err := btFilesMgr.CreateShareTask(torFile)
	if err != nil {
		log.Err(fmt.Sprintf("%s: %s", f, err.Error()))
		utils.CreateErrResp(w, &log, "Create share task fail.")
		return
	}

	// 3. 上传共享文件 bt 元数据
	peer_id := "1qaz2wsx3edc4rfv5tgb6yhn7ujm8ik9"
	serv := setting.AppSetting.GetTrackerServ()
	url := fmt.Sprintf("http://%s:%d/torrent?infohash=%s&peer_id=%s&port=%d",
		serv.Ip,
		serv.Port,
		infohash,
		peer_id,
		setting.AppSetting.GetBtServ().Port)
	log.Info(url)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(torFile)))
	if err != nil {
		utils.CreateErrResp(w, &log, fmt.Sprint("Create share task fail. %s", err.Error()))
		return
	}
	req.Host = serv.Ip
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		utils.CreateErrResp(w, &log, fmt.Sprint("Create share task fail. %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	result := map[string]interface{}{}
	utils.CreateSuccResp(w, &log, "Create share file task succ.", result)
}
