package nodetool

import (
	"crypto/md5"
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
				md5sum, _ := creator.calcFileMd5(
					path.Join(appSetting.GetAbResPath(), fileInfo.Name()))
				log.Printf("File info: name=%s, size=%d, md5sum=%s, mtime=%s\n",
					fileInfo.Name(),
					fileInfo.Size(),
					md5sum,
					fileInfo.ModTime().Format(time.RFC3339))
			}
		}
	}
	return []string{}, nil
}

func (creator *CreatorTorrent) calcFileMd5(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (creator *CreatorTorrent) File2TorrentFile(filePath string) error {
	return errors.New("create torrent file fail.")
}
