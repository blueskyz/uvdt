/*
	info hash 和 peer 管理结构
*/

package tracker

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
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
	exists, err := redis.Bool(rds.Do("EXISTS", infoHash))
	if err != nil && exists {
		return err
	}

	_, err = rds.Do("SET", infoHash, torrent)
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
	r.Close()

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
	torrent, err := redis.String(rds.Do("GET", infoHash))
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
		rows := DB.QueryRow(`Select infohash, torrent from infohash where 
							 infohash = ? limit 1`,
			infoHash)

		var infoHash string
		var torrent string
		err = rows.Scan(&infoHash, &torrent)
		if err != nil {
			return "", err
		}
		// 4. 找到 info hash 信息，插入缓存
		rds.Send("LPUSH", infoHash, torrent)
		rds.Flush()
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

	// 1. 从缓存查找 info hash 的 peer 信息
	peerKey := fmt.Sprintf("p_ih:%s", infoHash)
	peerList, err := redis.Values(rds.Do("hgetall", peerKey))
	if err != nil {
		return nil, err
	}

	// 2. 在缓存中找到 peer
	var peers []string
	if len(peerList) > 0 {
		find := false
		for _, v := range peerList {
			peer := string(v.([]byte))
			peers = append(peers, peer)
			if info.peer == peer {
				find = true
			}
		}

		if !find {
			rds.Do("hset", peerKey, info.peerId, info.peer)
		}
	} else {
		rds.Do("hset", peerKey, info.peerId, info.peer)

		// 3. 没有在缓存中找到，从数据库查找
		// 从数据库获取 info 信息
		rows, err := DB.Query(`Select peers from infohash where 
							   infohash = ? limit 1`,
			infoHash)
		rows.Close()
		if err != nil {
			return []string{}, err
		}
		count := 0
		var jsonPeers string
		for rows.Next() {
			err = rows.Scan(&jsonPeers)
			if err != nil {
				return []string{}, err
			}
			count += 1
		}

		// 4. 找到 info hash 信息，插入缓存
		if count > 0 {
			err = json.Unmarshal([]byte(jsonPeers), &peers)
			if err != nil {
				return []string{}, err
			}
			for _, peer := range peers {
				if peer != info.peer {
					rds.Send("hset", peerKey, info.peerId, peer)
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

	return peers, nil
}
