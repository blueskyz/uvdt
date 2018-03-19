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
// 路径配置，服务器配置，内存配置
type Setting struct {
	peerId   string
	logFile  string
	rootPath string // 上传，下载的资源文件路径

	httpServ    Serv
	btServ      Serv
	trackerServ Serv

	maxFileNum    uint // 并行管理的可以上传下载的文件数量，每个任务对应一个文件
	maxTaskNum    int  // 下载单个文件对应的协程数量
	maxMemPerTask uint // 每个上传下载任务可以使用的内存大小，单位：M
}

var AppSetting Setting

func init() {
	AppSetting = Setting{
		maxFileNum:    128,
		maxTaskNum:    32,
		maxMemPerTask: 32}
}

// root 目录
func (set *Setting) SetRootPath(rootPath string) {
	set.rootPath = rootPath
}

func (set *Setting) GetRootPath() string {
	return set.rootPath
}

// 设置日志文件路径
func (set *Setting) SetLogFile(logFile string) {
	set.logFile = logFile
}
func (set *Setting) GetLogFile() string {
	return set.logFile
}

func (set *Setting) GetMaxFileNum() uint {
	return set.maxFileNum
}

func (set *Setting) GetTaskNumForFile() int {
	return set.maxTaskNum
}

func (set *Setting) GetMaxMemPerFile() uint {
	return set.maxMemPerTask
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
	if len(ipport) != 2 {
		return Serv{}, errors.New(fmt.Sprintf("ip，port err: %s", value))
	}
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
