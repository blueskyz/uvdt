package nodetool

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blueskyz/uvdt/node-tool/setting"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

type CreatorTorrent struct {
}

func (creator *CreatorTorrent) ScanPath() ([]string, error) {
	appSetting := &setting.AppSetting
	fileList, err := filepath.Glob(path.Join(appSetting.GetAbResPath(), "*"))
	if err != nil {
		return []string{}, err
	} else {
		for _, file := range fileList {
			fileInfo, err := os.Stat(file)
			if err != nil {
				return []string{}, err
			}
			if !fileInfo.IsDir() {
				file_md5sum, fragments_md5sum, err := creator.calcFileMd5(
					path.Join(appSetting.GetAbResPath(), fileInfo.Name()))

				if err != nil {
					return []string{}, err
				}

				log.Printf("File info: name=%s, size=%d, md5sum=%s, mtime=%s\n",
					fileInfo.Name(),
					fileInfo.Size(),
					file_md5sum,
					fileInfo.ModTime().Format(time.RFC3339))

				c := make(map[string]interface{})
				c["file_name"] = fileInfo.Name()
				c["file_size"] = fileInfo.Size()
				c["file_md5"] = file_md5sum
				c["mtime"] = fileInfo.ModTime().UnixNano()
				c["file_fragments"] = fragments_md5sum
				torrent, err := json.Marshal(c)
				if err != nil {
					return []string{}, err
				}
				log.Printf("%s", torrent)
			}
		}
	}
	return []string{}, nil
}

func (creator *CreatorTorrent) calcFileMd5(filePath string) (string,
	[]string,
	error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", nil, err
	}
	file_md5sum := fmt.Sprintf("%x", h.Sum(nil))
	fragments_md5sum := []string{}

	return file_md5sum, fragments_md5sum, nil
}

func (creator *CreatorTorrent) File2TorrentFile(filePath string) error {
	return errors.New("create torrent file fail.")
}
