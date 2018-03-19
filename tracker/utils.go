/*
	通用组件
*/

package tracker

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/blueskyz/uvdt/logger"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

var (
	DB      *sql.DB
	RdsPool *redis.Pool
)

type peer struct {
	peer_id   string // 20 bytes
	address   string // ip
	port      int
	checktime int // 上一次有效检查时间
}

func InitDB(ipaddr string, user string, passwd string, dbname string) error {
	connStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4",
		user,
		passwd,
		ipaddr,
		dbname)
	dbconn, err := sql.Open("mysql", connStr)
	if err != nil {
		return err
	} else {
		DB = dbconn
	}
	dbconn.SetMaxIdleConns(0)
	return nil
}

func InitRedis(ipaddr string, passwd string, db string) {
	RdsPool = newPool(ipaddr, passwd, db)
}

func newPool(addr string, passwd string, db string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", passwd); err != nil {
				c.Close()
				return nil, err
			}
			if _, err := c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}
}

// 创建处理成功的返回内容
func CreateSuccResp(w http.ResponseWriter,
	log *logger.LogAgent,
	msg string,
	result map[string]interface{}) {

	log.Info(msg)
	retResp := map[string]interface{}{
		"status": 0,
		"msg":    msg,
		"result": result,
	}
	w.Header().Set("Content-type", "application/json")

	resp, err := json.Marshal(retResp)
	if err != nil {
		log.Err(fmt.Sprintf("Json serialize fail, ", err.Error()))
		errJsonResp := map[string]interface{}{
			"status": -1,
			"msg":    "Json serialize fail",
		}
		errJson, _ := json.Marshal(errJsonResp)
		w.Write(errJson)
	} else {
		w.Write(resp)
	}
}

// 创建处理错误的返回内容
func CreateErrResp(w http.ResponseWriter, log *logger.LogAgent, errMsg string) {
	log.Err(errMsg)
	errResp := map[string]interface{}{
		"status": -1,
		"msg":    errMsg,
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	resp, err := json.Marshal(errResp)
	if err != nil {
		log.Err(fmt.Sprintf("Json serialize fail, ", err.Error()))
		errJsonResp := map[string]interface{}{
			"status": -1,
			"msg":    "Json serialize fail",
		}
		errJson, _ := json.Marshal(errJsonResp)
		w.Write(errJson)
	} else {
		w.Write(resp)
	}
}

// 检查 hexdigest 字符串
func CheckHexdigest(value string, size int) bool {
	if len(value) != size {
		return false
	}
	value = strings.ToLower(value)
	pattern := fmt.Sprintf("[a-z0-9]{%d}", size)
	succ, _ := regexp.Match(pattern, []byte(value))
	// fmt.Printf("succ %s, %d, %v\n", value, size, succ)
	return succ
}

// 检查 ip 字符串
func CheckIP(sip string) bool {
	const ipReg = "\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}"
	reg, _ := regexp.Compile(ipReg)
	return reg.MatchString(sip)
}
