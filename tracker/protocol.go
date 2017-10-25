/*
	tracker bt http server 服务，提供 node peer 访问
*/
package tracker

import (
	"encoding/json"
)

func ParseBtProto(infoHash string, torrentContent string) (map[string]interface{}, error) {
	content := map[string]interface{}{}
	err := json.Unmarshal([]byte(torrentContent), &content)
	if err != nil {
		return nil, err
	}
	return content, err
}
