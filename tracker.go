/*
	tracker 服务
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/tracker"
	"github.com/blueskyz/uvdt/tracker/setting"
)

// 解析命令行参数配置
func ParseCmd() error {
	clusterList := flag.String("clusterip",
		"",
		"comma-separated ip list of tracker servers, ie 192.168.2.1:30083")
	btServ := flag.String("btserv",
		"0.0.0.0:80",
		"bt server ip and port")
	trackerServ := flag.String("trackerserv",
		"0.0.0.0:30080",
		"tracker server ip and port")
	logFile := flag.String("logfile",
		"/var/log/uvdt-trace.log",
		"log file")

	// 数据库配置
	dbHost := flag.String("db-host",
		"127.0.0.1:3306",
		"database host -- ip:port")
	dbUser := flag.String("user",
		"root",
		"database user")
	dbPasswd := flag.String("db-passwd",
		"123456",
		"database port")
	dbname := flag.String("dbname",
		"uvdt",
		"database name")

	// redis 配置
	redisHost := flag.String("redis-host",
		"127.0.0.1:6379",
		"redis host -- ip:port")
	redisPasswd := flag.String("redis-passwd",
		"",
		"redis auth")
	redisDB := flag.String("redis-db",
		"3",
		"redis database number")

	flag.Parse()

	// 打印服务参数
	log.Printf("cluster ip list: %s", *clusterList)
	log.Printf("bt server ip port: %s", *btServ)
	log.Printf("tracker server ip port: %s", *trackerServ)
	log.Printf("log file: %s", *logFile)

	log.Printf("database: host:%s, user: %s, passwd: ***, dbname: %s",
		*dbHost,
		*dbUser,
		*dbname)
	log.Printf("redis: host:%s, db: %s", *redisHost, *redisDB)

	// 创建配置对象
	AppSetting := &setting.AppSetting
	AppSetting.SetLogFile(*logFile)
	err := AppSetting.SetClusterList(*clusterList)
	if err == nil {
		err = AppSetting.SetBtServ(*btServ)
	}
	if err == nil {
		err = AppSetting.SetTraceServ(*trackerServ)
	}

	// 保存数据库、redis 配置
	AppSetting.SetDB(*dbHost, *dbUser, *dbPasswd, *dbname)
	AppSetting.SetRedis(*redisHost, *redisPasswd, *redisDB)

	return err
}

// 初始化设置
func init() {
	// 解析命令行参数
	err := ParseCmd()
	if err != nil {
		log.Printf("Error: %s", err)
		flag.Usage()
		os.Exit(-1)
	}

	// 设置日志
	logger.LoggerInit(setting.AppSetting.GetLogFile())
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	// 获取配置参数
	AppSetting := &setting.AppSetting
	dbHost, dbUser, dbPasswd, dbname := AppSetting.GetDB()
	redisHost, redisPasswd, redisDB := AppSetting.GetRedis()
	btServ := AppSetting.GetBtServ()
	trackerServ := AppSetting.GetTrackerServ()

	// 打印服务参数
	log.Info(fmt.Sprintf("bt server ip port: %s:%s", btServ.Ip, btServ.Port))
	log.Info(fmt.Sprintf("tracker server ip port: %s",
		trackerServ.Ip,
		trackerServ.Port))

	// log.Info(fmt.Sprintf("db: host:%s, user: %s, passwd: %s, dbname: %s",
	log.Info(fmt.Sprintf("db: host:%s, user: %s, passwd: ***, dbname: %s",
		dbHost,
		dbUser,
		// dbPasswd,
		dbname))
	log.Info(fmt.Sprintf("redis: host:%s, db: %s", redisHost, redisDB))

	// 设置数据库
	err = tracker.InitDB(dbHost, dbUser, dbPasswd, dbname)
	if err != nil {
		log.Err(fmt.Sprintf("db:%s", err.Error()))
	}

	// 设置数缓存
	tracker.InitRedis(redisHost, redisPasswd, redisDB)
}

func main() {
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	log.Info("Start tracker server ...")

	// 启动管理服务器
	go tracker.TrackerHttpServ()

	// 启动 bt http 服务
	tracker.BtHttpServ()
}
