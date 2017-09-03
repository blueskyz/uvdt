package nodetool

import (
	"errors"
	"github.com/blueskyz/uvdt/node-tool/setting"
	"log"
	"os"
	"path"
	"path/filepath"
)

func walkFunc(path string, info os.FileInfo, err error) error {
	log.Printf("file info: %v\n", info)
	return nil
}

func ScanPath() ([]string, error) {
	appSetting := &setting.AppSetting
	fileList, err := filepath.Glob(path.Join(appSetting.GetResPath(), "*"))
	if err != nil {
		return []string{}, nil
	} else {
		log.Printf("File info: %v\n", fileList)
	}
	return []string{}, nil
}

func File2TorrentFile(filePath string) error {
	return errors.New("create torrent file fail.")
}
