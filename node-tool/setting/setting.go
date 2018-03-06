/*
	配置管理，基本配置结构类型定义
*/
package setting

import (
	"errors"
	"fmt"
	"path"
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
// 保存要处理的资源设置
type Setting struct {
	logFile  string
	rootPath string // 上传，下载的资源文件路径

	resPath string
	resFile string

	torrentPath string

	trackerServ Serv
}

var AppSetting Setting

func init() {
}

// 设置日志文件路径
func (set *Setting) SetLogFile(logFile string) {
	set.logFile = logFile
}
func (set *Setting) GetLogFile() string {
	return set.logFile
}

// 设置项目 root 目录，所有配置文件，资源文件都在这个目录下
func (set *Setting) SetRootPath(rootPath string) {
	set.rootPath = rootPath
}

func (set *Setting) GetRootPath() string {
	return set.rootPath
}

// 设置扫描的目录，为了创建 infohash 和 torrentfile
func (set *Setting) SetResPath(resPath string) {
	set.resPath = resPath
}

func (set *Setting) GetResPath() string {
	return set.resPath
}

func (set *Setting) GetAbResPath() string {
	return path.Join(set.rootPath, set.resPath)
}

// 设置 torrent 文件路径
func (set *Setting) SetTorrentPath(torrentPath string) {
	set.torrentPath = torrentPath
}

func (set *Setting) GetAbTorrentPath() string {
	return path.Join(set.rootPath, "share", set.torrentPath)
}

// 设置要共享的资源文件，为了创建 infohash 和 torrentfile
func (set *Setting) SetResFile(resFile string) {
	set.resFile = resFile
}

func (set *Setting) GetResFile() string {
	return set.resFile
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
