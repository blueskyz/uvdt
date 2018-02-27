/*
	http server 服务，提供管理访问
*/
package nodeserv

import (
	"fmt"
	"net/http"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/node-serv/setting"
	"github.com/blueskyz/uvdt/utils"
)

var filesMgr *FilesManager

func HttpServ(filesManager *FilesManager) error {
	log := logger.NewAgent()
	defer log.EndLog()

	filesMgr = filesManager

	// 设置  http server 路由
	//***********************************************************************
	// 页面路由
	HttpServMux := http.NewServeMux()
	HttpServMux.HandleFunc("/hello", httpHelloHandler)
	HttpServMux.HandleFunc("/", httpHandler)

	//***********************************************************************
	// api 接口
	HttpServMux.HandleFunc("/api/stats", apiStatsHandler)

	// 上传分享的 tor 文件
	HttpServMux.HandleFunc("/api/upload", httpHandler)

	// 添加下载任务
	HttpServMux.HandleFunc("/api/download", httpHandler)

	httpServ := setting.AppSetting.GetHttpServ()
	log.Info(fmt.Sprintf("%s:%d", httpServ.Ip, httpServ.Port))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d",
		httpServ.Ip,
		httpServ.Port),
		HttpServMux)
	fmt.Println("why ...")
	if err != nil {
		log.Err(err.Error())
	}
	return err
}

func httpHelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello http serv")
}

/*
 * 管理访问页面
 */
func httpHandler(w http.ResponseWriter, r *http.Request) {
}

/*
 * 管理 api
 */
func apiStatsHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.NewAgent()
	defer log.EndLog()

	// 输出服务器状态信息
	stats, err := filesMgr.GetStats()
	if err == nil {
		utils.CreateSuccResp(w, &log, "Get Stats succ", stats)
	} else {
		utils.CreateErrResp(w, &log, "Can't show stats")
	}

	/*
		w.Write([]byte(showFilesList))
	*/
}
