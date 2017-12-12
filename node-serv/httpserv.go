/*
	http server 服务，提供管理访问
*/
package nodeserv

import (
	"fmt"
	"net/http"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/node-serv/setting"
)

var filesMgr *FilesManager

func HttpServ(filesManager *FilesManager) error {
	log := logger.NewAgent()
	defer log.EndLog()

	filesMgr = filesManager

	// 设置  http server 路由
	HttpServMux := http.NewServeMux()
	HttpServMux.HandleFunc("/hello", httpHelloHandler)
	HttpServMux.HandleFunc("/", httpHandler)

	// 上传
	HttpServMux.HandleFunc("/api/upload", httpHandler)

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
 * 管理访问
 */
func httpHandler(w http.ResponseWriter, r *http.Request) {
	showDownLoadStat := fmt.Sprintf("http filesMgr maxFileNum=%d, currentNum=%d",
		filesMgr.GetMaxFileNum(),
		filesMgr.GetCurrentFileNum())
	w.Write([]byte(showDownLoadStat))
}
