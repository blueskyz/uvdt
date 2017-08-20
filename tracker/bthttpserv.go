/*
	tracker bt http server 服务，提供 node peer 访问
*/
package tracker

import (
	"fmt"
	"net/http"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/tracker/setting"
)

func BtHttpServ() {
	// 设置 bt http server 路由
	btHttpServMux := http.NewServeMux()
	btHttpServMux.HandleFunc("/hello", btHelloHandler)
	btHttpServMux.HandleFunc("/", btHandler)

	btServ := setting.AppSetting.GetBtServ()
	logger.Info(fmt.Sprintf("%s:%d", btServ.Ip, btServ.Port))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", btServ.Ip, btServ.Port), btHttpServMux)
	if err != nil {
		logger.Err(err.Error())
	}
}

func btHelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello bt http serv")
}

func btHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "bt http serv")
}
