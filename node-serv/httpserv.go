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

func HttpServ() {
	log := logger.NewAgent()
	defer log.EndLog()

	// 设置  http server 路由
	HttpServMux := http.NewServeMux()
	HttpServMux.HandleFunc("/hello", httpHelloHandler)
	HttpServMux.HandleFunc("/", httpHandler)

	// 上传
	HttpServMux.HandleFunc("/api/upload", httpHandler)

	httpServ := setting.AppSetting.GetHttpServ()
	log.Info(fmt.Sprintf("%s:%d", httpServ.Ip, httpServ.Port))
	fmt.Println("why ...")
	err := http.ListenAndServe(fmt.Sprintf("%s:%d",
		httpServ.Ip,
		httpServ.Port),
		HttpServMux)
	if err != nil {
		log.Err(err.Error())
	}
}

func httpHelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello http serv")
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "http serv")
}
