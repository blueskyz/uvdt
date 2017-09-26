package nodetool

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blueskyz/uvdt/node-tool/setting"
	"io"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	// "time"
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
				fileMd5, partsMd5, err := creator.calcFileMd5(
					path.Join(appSetting.GetAbResPath(), fileInfo.Name()))

				if err != nil {
					return []string{}, err
				}

				/*
					log.Printf("File info: name=%s, size=%d, md5sum=%s, mtime=%s\n",
						fileInfo.Name(),
						fileInfo.Size(),
						fileMd5,
						fileInfo.ModTime().Format(time.RFC3339))
				*/

				c := make(map[string]interface{})
				c["file_name"] = fileInfo.Name()
				c["file_size"] = fileInfo.Size()
				c["file_md5"] = fileMd5
				c["mtime"] = fileInfo.ModTime().UnixNano()
				c["file_parts"] = partsMd5
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
	fileMd5 := fmt.Sprintf("%x", h.Sum(nil))

	// 计算 512KB 一个分片的 md5
	f.Seek(0, os.SEEK_SET)
	fileInfo, _ := f.Stat()
	fileSize := fileInfo.Size()
	const fileChunk = 1 * (1 << 19)
	floatChunk := float64(fileSize) / float64(fileChunk)
	totalPartsNum := uint64(math.Ceil(floatChunk))
	partsMd5 := []string{}
	for i := uint64(0); i < totalPartsNum; i++ {
		leftSize := float64(fileSize - int64(i*fileChunk))
		partSize := uint64(math.Min(fileChunk, leftSize))
		partBuffer := make([]byte, partSize)
		f.Read(partBuffer)
		partMd5 := fmt.Sprintf("%x", md5.Sum(partBuffer))
		partsMd5 = append(partsMd5, partMd5)
	}

	return fileMd5, partsMd5, nil
}

func (creator *CreatorTorrent) File2TorrentFile(filePath string) error {
	return errors.New("create torrent file fail.")
}
