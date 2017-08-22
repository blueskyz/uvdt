/*
	通用组件
*/

package tracker

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

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

func InitDB(ipaddr string, user string, passwd string) error {
	dbconn, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s)/uvdt?charset=utf8mb4", user, passwd, ipaddr))
	if err != nil {
		return err
	} else {
		DB = dbconn
	}
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

// 创建处理错误的返回内容
func CreateErrResp(statusCode int, msg string) []byte {
	errResp := map[string]interface{}{
		"status": statusCode,
		"msg":    msg,
	}
	resp, _ := json.Marshal(errResp)
	return resp
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
