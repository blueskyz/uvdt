/*
	info hash 和 peer 管理结构
*/

package tracker

type InfoHash struct {
	infoHash string
	name     string
	peers    string
}

// 1. 检查 peer 是否是公网 ip
// 2. 检查缓存中是否存在 info hash 不存在则添加 info hash 和 peer 到数据库，并更新缓存
// 3. 如果存在，则添加 peer 到 info hash 结构
func (info *InfoHash) GetInfoHash(infoHash string) error {
	// 从缓存获取 info 信息
	rds := RdsPool.Get()
	defer rds.Close()

	// 从数据库获取 info 信息
	rows, err := DB.Query("Select * from torrent where infohash = ?", infoHash)
	if err != nil {
		return err
	}
	for rows.Next() {
		err = rows.Scan(&info.infoHash, &info.name, &info.peers)
		if err != nil {
			return err
		}
	}
	return nil
}

func (peer *peer) addPeer() {
}
