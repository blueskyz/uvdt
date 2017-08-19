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
	ip   string
	port int
}

// 配置类型
type Setting struct {
	logFile      string
	btServ       Serv
	trackerServ  Serv
	clusterServs []Serv
}

// 获取逗号分割的字符串参数
func (set *Setting) SetClusterList(value string) error {
	if len(value) == 0 {
		return errors.New(fmt.Sprintf("Cluster list is empty"))
	}
	for _, curValue := range strings.Split(value, ",") {
		ipport := strings.Split(curValue, ":")
		port, err := strconv.Atoi(ipport[1])
		if err != nil {
			return errors.New(fmt.Sprintf("Cluster list ip err: %s", curValue))
		}
		curSet := Serv{
			ip:   ipport[0],
			port: port,
		}
		set.clusterServs = append(set.clusterServs, curSet)
	}
	return nil
}

// 设置日志文件路径
func (set *Setting) SetLogFile(logFile string) {
	set.logFile = logFile
}
func (set *Setting) GetLogFile() string {
	return set.logFile
}

// 设置 bt server
func (set *Setting) SetBtServ(value string) error {
	btServ, err := str2Serv(value)
	if err == nil {
		set.btServ = btServ
	}
	return err
}

// 设置 trace server
func (set *Setting) SetTraceServ(value string) error {
	trackerServ, err := str2Serv(value)
	if err == nil {
		set.trackerServ = trackerServ
	}
	return err
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
		ip:   ipport[0],
		port: port,
	}
	return serv, nil
}
