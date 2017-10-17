/*
	tracker http server 服务，提供管理访问
*/
package tracker

import (
	"fmt"
	"net/http"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/tracker/setting"
)

func TrackerHttpServ() {
	log := logger.NewAgent()
	defer log.EndLog()

	// 设置 tracker http server 路由
	trackerHttpServMux := http.NewServeMux()
	trackerHttpServMux.HandleFunc("/hello", trackerHelloHandler)
	trackerHttpServMux.HandleFunc("/", trackerHandler)

	trackerServ := setting.AppSetting.GetTrackerServ()
	log.Info(fmt.Sprintf("%s:%d", trackerServ.Ip, trackerServ.Port))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d",
		trackerServ.Ip,
		trackerServ.Port),
		trackerHttpServMux)
	if err != nil {
		log.Err(err.Error())
	}
}

func trackerHelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello tracker http serv")
}

func trackerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Start tracker http serv ...")
}
