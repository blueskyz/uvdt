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
	fmt.Fprintf(w, "http serv hello")
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
	fmt.Fprintf(w, "http serv api create share resource")

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
	f, err := os.OpenFile(torrent_file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Err(fmt.Sprintf("%s: %s", f, err.Error()))
		return
	}
	defer f.Close()

	torFile := make([]byte, 1024<<10)
	count, err := f.Read(torFile)
	if err != nil && err != io.EOF {
		log.Err(fmt.Sprintf("Read tor data fail, %s", torrent_file))
		return
	}
	torFile = torFile[:count]

	// 2. 从 share 目录找到共享的文件

	// 3. 创建本地共享文件
	btFilesMgr.CreateShareTask(string(torFile))

	// 4. 上传共享文件 bt 元数据
}
