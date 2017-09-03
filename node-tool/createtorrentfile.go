package nodetool

import (
	"errors"
	"github.com/blueskyz/uvdt/node-tool/setting"
	// "log"
	"os"
	"path"
	"path/filepath"
	// "time"
)

type CreatorTorrent struct {
}

func (creator *CreatorTorrent) ScanPath() ([]string, error) {
	appSetting := &setting.AppSetting
	fileList, err := filepath.Glob(path.Join(appSetting.GetResPath(), "*"))
	if err != nil {
		return []string{}, err
	} else {
		for _, file := range fileList {
			fileInfo, err := os.Stat(file)
			if err != nil {
				return []string{}, err
			}
			if !fileInfo.IsDir() {
				/*
					log.Printf("File info: name=%s, size=%d, mtime=%s\n",
						fileInfo.Name(),
						fileInfo.Size(),
						fileInfo.ModTime().Format(time.UnixDate))
				*/
			}
		}
	}
	return []string{}, nil
}

func (creator *CreatorTorrent) File2TorrentFile(filePath string) error {
	return errors.New("create torrent file fail.")
}
