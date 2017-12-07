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
	HttpBtServMux := http.NewServeMux()
	HttpBtServMux.HandleFunc("/hello", httpBtHelloHandler)

	// 上传
	HttpBtServMux.HandleFunc("/api/download", httpBtHandler)

	httpBtServ := setting.AppSetting.GetBtServ()
	log.Info(fmt.Sprintf("%s:%d", httpBtServ.Ip, httpBtServ.Port))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d",
		httpBtServ.Ip,
		httpBtServ.Port),
		HttpBtServMux)
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
func httpBtHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "download http serv")
}
