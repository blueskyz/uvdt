/*
	配置管理，基本配置结构类型定义
*/
package setting

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// 服务类型 ip, port
type Serv struct {
	Ip   string
	Port int
}

// 配置类型
type Setting struct {
	logFile     string
	httpServ    Serv
	btServ      Serv
	trackerServ Serv
}

var AppSetting Setting

func init() {
	AppSetting = Setting{}
}

// 设置日志文件路径
func (set *Setting) SetLogFile(logFile string) {
	set.logFile = logFile
}
func (set *Setting) GetLogFile() string {
	return set.logFile
}

// 设置 http server
func (set *Setting) SetHttpServ(value string) error {
	httpServ, err := str2Serv(value)
	if err == nil {
		set.httpServ = httpServ
	}
	return err
}

func (set *Setting) GetHttpServ() Serv {
	return set.httpServ
}

// 设置 bt server
func (set *Setting) SetBtServ(value string) error {
	btServ, err := str2Serv(value)
	if err == nil {
		set.btServ = btServ
	}
	return err
}

func (set *Setting) GetBtServ() Serv {
	return set.btServ
}

// 设置 trace server
func (set *Setting) SetTraceServ(value string) error {
	trackerServ, err := str2Serv(value)
	if err == nil {
		set.trackerServ = trackerServ
	}
	return err
}

func (set *Setting) GetTrackerServ() Serv {
	return set.trackerServ
}

// 获取 Serv 对象
func str2Serv(value string) (Serv, error) {
	if len(value) == 0 {
		return Serv{}, errors.New(fmt.Sprintf("Serv convert argument empty"))
	}
	ipport := strings.Split(value, ":")
	port, err := strconv.Atoi(ipport[1])
	if err != nil {
		return Serv{}, errors.New(fmt.Sprintf("ip，port err: %s", value))
	}
	serv := Serv{
		Ip:   ipport[0],
		Port: port,
	}
	return serv, nil
}
