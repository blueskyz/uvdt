/*
	info hash 和 peer 管理结构
*/

package tracker

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"strings"
)

type Torrent struct {
	infoHash string
	name     string
	peer     string // peer_id:ip:port
	peerId   string
	status   int
}

func (info *Torrent) AddTorrent(infoHash string, torrent string) error {
	// 0. 验证用户

	// 从缓存获取 info 信息
	rds := RdsPool.Get()
	defer rds.Close()

	// 1. 从缓存查找 info hash 信息
	torKey := fmt.Sprintf("tor:%s", infoHash)
	exists, err := redis.Bool(rds.Do("EXISTS", torKey))
	if err != nil && exists {
		return err
	}

	_, err = rds.Do("SET", torKey, torrent)
	if err != nil {
		return err
	}

	// 2. 保存到数据库
	r, err := DB.Query(`insert into infohash (infohash, ctime, torrent) 
					   values (?, unix_timestamp(), ?)`,
		infoHash,
		torrent)
	if err != nil {
		return err
	}
	defer r.Close()

	return nil
}

// 1. 检查缓存中是否存在 info hash 不存在则添加 info hash 到数据库，并更新缓存
// 2. 如果存在，则添加 peer 到 info hash 结构
func (info *Torrent) GetTorrent(infoHash string) (string, error) {
	// 0. 验证用户

	// 从缓存获取 info 信息
	rds := RdsPool.Get()
	defer rds.Close()

	// 1. 从缓存查找 info hash 信息
	torKey := fmt.Sprintf("tor:%s", infoHash)
	torrent, err := redis.String(rds.Do("get", torKey))
	if err != nil {
		if err.Error() != "redigo: nil returned" {
			return "", err
		}
	}

	// 2. 在缓存中找到 torrent
	if len(torrent) > 0 {
		return torrent, nil
	} else {
		// 3. 没有在缓存中找到，从数据库查找
		// 从数据库获取 info 信息
		rows := DB.QueryRow(`select infohash, torrent from infohash where 
							 infohash = ? limit 1`,
			infoHash)

		var infoHash string
		var torrent string
		err = rows.Scan(&infoHash, &torrent)
		switch {
		case err == sql.ErrNoRows:
			return "", nil
		case err != nil:
			return "", err
		}

		// 4. 找到 info hash 信息，插入缓存
		rds.Do("set", torKey, torrent)
		return torrent, nil
	}

	return "", nil
}

// 1. 检查 peer 是否是公网 ip
// 2. 检查缓存中是否存在 info hash 不存在则添加 info hash 和 peer 到数据库，并更新缓存
// 3. 如果存在，则添加 peer 到 info hash 结构
func (info *Torrent) GetPeers(infoHash string) ([]string, error) {
	// 从缓存获取 info 信息
	rds := RdsPool.Get()
	defer rds.Close()

	// 1. 根据 peerId 更新缓存的 peer 信息
	peerCurKey := fmt.Sprintf("pik:%s", info.peerId)
	peerInfo, err := redis.String(rds.Do("get", peerCurKey))
	if err != nil {
		if err.Error() != "redigo: nil returned" {
			return nil, err
		}
	}
	if peerInfo != info.peer {
		rds.Do("set", peerCurKey, info.peer)
	}

	// 2. 从缓存查找 info hash 的 peer 信息
	infoKey := fmt.Sprintf("ih:%s", infoHash)
	peersMap, err := redis.StringMap(rds.Do("zrange", infoKey, 0, 100, "withscores"))
	if err != nil {
		return nil, err
	}

	// 3. 在缓存中找到 peer
	var peerIds []string
	if len(peersMap) > 0 {
		find := false
		for peerIdKey := range peersMap {
			if peerCurKey == peerIdKey {
				find = true
			} else {
				peerIds = append(peerIds, peerIdKey)
			}
		}

		if !find {
			rds.Do("zadd", infoKey, 1, peerCurKey)
		}
	} else {
		rds.Do("zadd", infoKey, 1, peerCurKey)

		// 4. 没有在缓存中找到，从数据库查找
		// 从数据库获取 info 信息
		rows, err := DB.Query(`select peers from infohash where
							   infohash = ? limit 1`,
			infoHash)
		if err != nil {
			return []string{}, err
		}
		defer rows.Close()

		count := 0
		var jsonPeers string
		for rows.Next() {
			err = rows.Scan(&jsonPeers)
			if err != nil {
				return []string{}, err
			}
			count += 1
		}

		// 4. 找到数据库中的 info hash 信息，插入缓存
		if count > 0 {
			var peers []string
			err = json.Unmarshal([]byte(jsonPeers), &peers)
			if err != nil {
				return []string{}, err
			}
			for _, peer := range peers {
				peerId := strings.Split(peer, ":")[0]
				if peerId != info.peerId {
					peerIdKey := fmt.Sprintf("pik:%s", peerId)
					rds.Send("set", peerIdKey, peer)
					rds.Send("zadd", infoKey, 1, peerIdKey)
					peerIds = append(peerIds, peerIdKey)
				}
			}
			rds.Flush()
		} else {
			// 5. 没有找到 info hash 信息，保存 info hash 信息到数据库
			peers_value, err := json.Marshal([]string{info.peer})
			stmp, err := DB.Prepare(`update infohash
			 						 set name=?,
									 peers=?,
									 ctime=unix_timestamp(),
									 mtime=unix_timestamp() 
									 where infohash=?`)
			_, err = stmp.Exec(info.name, peers_value, infoHash)
			if err != nil {
				return []string{}, err
			}
		}
	}

	var args []interface{}
	for _, k := range peerIds {
		args = append(args, k)
	}

	if len(args) <= 0 {
		return []string{}, nil
	}

	peers, err := redis.Strings(rds.Do("mget", args...))
	if err != nil {
		return []string{}, err
	}

	return peers, nil
}
