/*
	bt http server 服务，提供下载服务
*/
package nodeserv

import (
	"fmt"
	"net/http"

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

	// 1. 从 share 目录找到共享的文件

	// 2. 创建本地共享文件
	btFilesMgr.CreateShareTask(fileMd5)

	// 3. 上传共享文件 bt 元数据
}
