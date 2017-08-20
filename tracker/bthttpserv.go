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
	log := logger.NewAgent()
	// 设置 bt http server 路由
	btHttpServMux := http.NewServeMux()
	btHttpServMux.HandleFunc("/hello", btHelloHandler)
	btHttpServMux.HandleFunc("/", btHandler)

	btServ := setting.AppSetting.GetBtServ()
	log.Info(fmt.Sprintf("Init %s:%d", btServ.Ip, btServ.Port))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", btServ.Ip, btServ.Port), btHttpServMux)
	if err != nil {
		log.Err("init, " + err.Error())
	}
}

func btHelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello bt http serv")
}

func btHandler(w http.ResponseWriter, r *http.Request) {
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	log.Info("This is Info 1")
	fmt.Fprintf(w, fmt.Sprintf("bt http serv %s", r.RequestURI))
	log.Info("This is Info 1")
	log.Err("This is Err 1")
	log.Err("This is Err 1")
}
