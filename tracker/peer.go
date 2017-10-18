/*
	info hash 和 peer 管理结构
*/

package tracker

import (
	"encoding/json"
	_ "fmt"
	"github.com/garyburd/redigo/redis"
)

type Torrent struct {
	infoHash string
	name     string
	peer     string // peer_id:ip:port
	status   int
}

func (info *Torrent) AddTorrent(infoHash string, torrent string) error {
	return nil
}

func (info *Torrent) GetTorrent(infoHash string) (string, error) {
	return "", nil
}

// 1. 检查 peer 是否是公网 ip
// 2. 检查缓存中是否存在 info hash 不存在则添加 info hash 和 peer 到数据库，并更新缓存
// 3. 如果存在，则添加 peer 到 info hash 结构
func (info *Torrent) GetPeers(infoHash string) ([]string, error) {
	// 从缓存获取 info 信息
	rds := RdsPool.Get()
	defer rds.Close()

	var peers []string
	// 1. 从缓存查找 info hash 的 peer 信息
	peerList, err := redis.Values(rds.Do("LRANGE", infoHash, "0", "100"))
	if err != nil {
		return nil, err
	}

	// 2. 在缓存中找到 peer
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
			rds.Do("LPUSH", infoHash, info.peer)
		}
	} else {
		rds.Do("LPUSH", infoHash, info.peer)

		// 3. 没有在缓存中找到，从数据库查找
		// 从数据库获取 info 信息
		rows, err := DB.Query(`Select peers from infohash where 
							   infohash = ? limit 1`,
			infoHash)
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
			find := false
			for _, peer := range peers {
				rds.Send("LPUSH", infoHash, peer)
				if peer == info.peer {
					find = true
				}
			}
			if !find {
				rds.Send("LPUSH", infoHash, info.peer)
			}
			rds.Flush()
		} else {
			// 5. 没有找到 info hash 信息，保存 info hash 信息到数据库
			peers_value, err := json.Marshal([]string{info.peer})
			stmp, err := DB.Prepare(`insert into 
									 infohash(infohash,
											  name,
											  peers,
											  ctime,
											  mtime)
									 values(?,
											?,
											?,
											unix_timestamp(),
											unix_timestamp())`)
			_, err = stmp.Exec(infoHash, info.name, peers_value)
			if err != nil {
				return []string{}, err
			}
		}
	}

	return peers, nil
}
