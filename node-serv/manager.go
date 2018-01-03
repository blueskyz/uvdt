/*
	manager server 服务，提供管理访问
*/

package nodeserv

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/node-serv/setting"
)

func CreateFilesMgr() (*FilesManager, error) {

	filesMgr := &FilesManager{
		maxFileNum: setting.AppSetting.GetMaxFileNum(),
		lock:       sync.RWMutex{},
	}

	// 1. 加载配置数据库
	err := filesMgr.LoadDB()
	if err != nil {
		return nil, err
	}

	// FilesManager{
	return filesMgr, nil
}

/*
 * 上传，下载文件管理
 */
type FilesManager struct {
	version    string
	maxFileNum uint
	lock       sync.RWMutex

	fileTasksMgr []FileTasksMgr
}

func (filesMgr *FilesManager) GetVersion() string {
	return filesMgr.version
}

func (filesMgr *FilesManager) GetMaxFileNum() uint {
	return filesMgr.maxFileNum
}

func (filesMgr *FilesManager) GetFileTasksMgr() []FileTasksMgr {
	return filesMgr.fileTasksMgr
}

func (filesMgr *FilesManager) GetCurrentFileNum() int {
	return len(filesMgr.fileTasksMgr)
}

func (filesMgr *FilesManager) GetRootPath() string {
	return setting.AppSetting.GetRootPath()
}

func (filesMgr *FilesManager) LoadDB() error {
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	metaPath := path.Join(setting.AppSetting.GetRootPath(), ".meta")

	// 1. 当 root path 不存在时创建目录
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create meta path, %s", metaPath))
		os.MkdirAll(metaPath, os.ModeDir|os.ModePerm)
	}

	// 2. 检查文件元数据，没有则创建
	jsonMetaFile := path.Join(metaPath, "meta.dat")
	if _, err := os.Stat(jsonMetaFile); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create meta data, %s", jsonMetaFile))
		blob := `{"version": "v1.0", "fileslist": ["test.txt"]}`
		f, err := os.OpenFile(jsonMetaFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Err(fmt.Sprintf("Create meta data fail, %s", jsonMetaFile))
			return err
		}
		f.Write([]byte(blob))
		f.Close()
	}

	// 3. 从 json 文件中加载共享的文件元数据
	f, err := os.Open(jsonMetaFile)
	if err != nil {
		log.Err(fmt.Sprintf("Open meta data fail, %s", jsonMetaFile))
		return err
	}
	defer f.Close()

	metaData := make([]byte, 1024<<10)
	count, err := f.Read(metaData)
	if err != nil && err != io.EOF {
		log.Err(fmt.Sprintf("Read meta data fail, %s", jsonMetaFile))
		return err
	}
	metaData = metaData[:count]

	// 4. 解析元数据
	jsonMeta := make(map[string]interface{})
	if err := json.Unmarshal(metaData, &jsonMeta); err != nil {
		log.Err(fmt.Sprintf("Parse json meta data fail, %s", jsonMetaFile))
		return err
	}
	filesMgr.version = jsonMeta["version"].(string)

	// 5. 创建 下载/共享 的文件管理器
	filesList := jsonMeta["fileslist"].([]interface{})
	for _, v := range filesList {
		fileTasksMgr := FileTasksMgr{}
		fileTasksMgr.Start(int(setting.AppSetting.GetTaskNumForFile()), v.(string))
		filesMgr.fileTasksMgr = append(filesMgr.fileTasksMgr, fileTasksMgr)
	}

	return nil
}

func (filesMgr *FilesManager) GetStats() (map[string]interface{}, error) {
	// lock
	filesMgr.lock.RLock()

	stats := map[string]interface{}{
		"version":      filesMgr.GetVersion(),
		"root_path":    filesMgr.GetRootPath(),
		"max_file_num": filesMgr.GetMaxFileNum(),
		"current_num":  filesMgr.GetCurrentFileNum(),
	}

	// 输出共享的文件列表
	/*
		fileTasksMgr := filesMgr.GetFileTasksMgr()
		var showFilesList []interface{}
		for _, v := range fileTasksMgr {
			showFilesList += v.fileMeta.GetFileTasksStats()
		}
	*/
	// unlock
	filesMgr.lock.RUnlock()

	return stats, nil
}
