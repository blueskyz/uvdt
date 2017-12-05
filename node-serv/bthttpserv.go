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

func BtHttpServ() {
	log := logger.NewAgent()
	defer log.EndLog()

	// 设置  http server 路由
	HttpDownloadServMux := http.NewServeMux()
	HttpDownloadServMux.HandleFunc("/hello", httpBtHelloHandler)

	// 上传
	HttpDownloadServMux.HandleFunc("/api/download", downloadHandler)

	httpDownloadServ := setting.AppSetting.GetBtServ()
	log.Info(fmt.Sprintf("%s:%d", httpDownloadServ.Ip, httpDownloadServ.Port))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d",
		httpDownloadServ.Ip,
		httpDownloadServ.Port),
		HttpDownloadServMux)
	if err != nil {
		log.Err(err.Error())
	}
}

func httpBtHelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello http download serv")
}

/*
 * 下载资源
 */
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "download http serv")
}
