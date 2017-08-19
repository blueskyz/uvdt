/*
	配置管理，基本配置结构类型定义
*/
package setting

import (
	"errors"
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
	btServ      Serv
	trackerServ Serv
	clusterip   []Serv
}

// 获取逗号分割的字符串参数
func (set *Setting) SetClusterList(value string) error {
	if len(value) > 0 {
		return errors.New("参数为空")
	}
	for _, curValue := range strings.Split(value, ",") {
		ipport := strings.Split(curValue, ":")
		port, err := strconv.Atoi(ipport[1])
		curSet := Serv{
			ip:   ipport[0],
			port: port,
		}
		*set = append(*set, curSet)
	}
	return nil
}
