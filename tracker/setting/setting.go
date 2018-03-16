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
	logFile      string
	btServ       Serv
	trackerServ  Serv
	clusterServs []Serv

	dbHost   string
	dbUser   string
	dbPasswd string
	dbname   string

	redisHost   string
	redisPasswd string
	redisDB     string
}

var AppSetting Setting

func init() {
	AppSetting = Setting{}
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
			Ip:   ipport[0],
			Port: port,
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

// 设置数据库
func (set *Setting) SetDB(dbHost string,
	dbUser string,
	dbPasswd string,
	dbname string) error {

	set.dbHost = dbHost
	set.dbUser = dbUser
	set.dbPasswd = dbPasswd
	set.dbname = dbname

	return nil
}

func (set *Setting) GetDB() (string, string, string, string) {
	return set.dbHost, set.dbUser, set.dbPasswd, set.dbname
}

// 设置 redis
func (set *Setting) SetRedis(redisHost string,
	redisPasswd string,
	redisDB string) error {

	set.redisHost = redisHost
	set.redisPasswd = redisPasswd
	set.redisDB = redisDB

	return nil
}

func (set *Setting) GetRedis() (string, string, string) {
	return set.redisHost, set.redisPasswd, set.redisDB
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
