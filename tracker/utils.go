/*
	通用组件
*/

package tracker

import (
	"database/sql"
	_ "encoding/json"
	"fmt"
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
		fmt.Sprintf("%s:%s@tcp(%s)/uvdt?charset=utf-8", user, passwd, ipaddr))
	if err != nil {
		return err
	} else {
		DB = dbconn
	}
	return nil
}

func InitRedis(ipaddr string, db string) {
	RdsPool = newPool(ipaddr, db)
}

func newPool(addr string, db string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
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
