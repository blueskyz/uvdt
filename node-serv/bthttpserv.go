/*
	bt http server 服务，提供下载服务
*/
package nodeserv

import (
	"fmt"
	"net/http"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/node-serv/setting"
)

var btFilesMgr *FilesManager

func BtHttpServ(filesManager *FilesManager) error {
	log := logger.NewAgent()
	defer log.EndLog()

	btFilesMgr = filesManager

	// 设置  http server 路由
	HttpBtServMux := http.NewServeMux()
	HttpBtServMux.HandleFunc("/hello", httpBtHelloHandler)

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
