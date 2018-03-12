/*
	manager server 服务，提供管理访问
*/

package nodeserv

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/node-serv/setting"
)

/*
 * 上传，下载文件管理
 */
type FilesManager struct {
	version    string
	maxFileNum uint
	lock       sync.RWMutex

	fileTasksMgr []FileTasksMgr
}

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

func (filesMgr *FilesManager) CreateShareTask(torrent []byte) (string,
	string, error) {
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	// lock
	filesMgr.lock.RLock()
	// unlock
	defer filesMgr.lock.RUnlock()

	fileTasksMgr := FileTasksMgr{lock: sync.RWMutex{}}
	filesMgr.fileTasksMgr = append(filesMgr.fileTasksMgr, fileTasksMgr)
	filename, fileMd5, err := fileTasksMgr.CreateShareFile(torrent)
	if err != nil {
		log.Err(fmt.Sprintf("Create share file fail, %s", err.Error()))
		return "", "", err
	}

	err = addToUvdtData(filename, fileMd5, "share")
	if err != nil {
		log.Err(fmt.Sprintf("Add to uvdt data fail, %s", err.Error()))
		return "", "", err
	}
	log.Info(fmt.Sprintf("Task %s[%s] started, state: %s",
		filename,
		fileMd5,
		"share"))

	return filename, fileMd5, nil
}

func addToUvdtData(filename string, md5 string, filepath string) error {
	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	// 1. 从 json 文件中加载共享的文件元数据
	uvdtRootPath := path.Join(setting.AppSetting.GetRootPath(), ".uvdt")
	uvdtJsonDataFile := path.Join(uvdtRootPath, "uvdt.dat")
	f, err := os.Open(uvdtJsonDataFile)
	if err != nil {
		log.Err(fmt.Sprintf("Open uvdt json data fail, %s", err.Error()))
		return err
	}
	defer f.Close()

	uvdtData := make([]byte, 1024<<10)
	count, err := f.Read(uvdtData)
	if err != nil && err != io.EOF {
		log.Err(fmt.Sprintf("Read uvdt data fail, %s", err.Error()))
		return err
	}
	uvdtData = uvdtData[:count]

	// 2. 解析元数据
	jsonMeta := make(map[string]interface{})
	if err := json.Unmarshal(uvdtData, &jsonMeta); err != nil {
		log.Err(fmt.Sprintf("Parse json uvdt data fail, %s", err.Error()))
		return err
	}
	filesMgr.version = jsonMeta["version"].(string)

	// 3. 添加文件数据到字典
	/*
		blob := `{"version": "v1.0",
		"fileslist": [{"filename": "test.txt", "path":"downloads", "md5": "xxx"}]}`
	*/
	filesList := jsonMeta["fileslist"].([]interface{})
	helpMd5Map := make(map[string]int)
	for _, i := range filesList {
		helpMd5Map[i.(map[string]interface{})["md5"].(string)] = 1
	}

	if _, ok := helpMd5Map[md5]; ok {
		return errors.New(fmt.Sprintf("torrent file exist, %s", md5))
	}

	fileInfo := make(map[string]interface{})
	fileInfo["filename"] = filename
	fileInfo["path"] = filepath
	fileInfo["md5"] = md5
	filesList = append(filesList, fileInfo)
	jsonMeta["fileslist"] = filesList

	// 4. 保存共享文件信息
	metaData, err := json.Marshal(jsonMeta)
	if err != nil {
		return err
	}

	jsonFile, err := os.OpenFile(uvdtJsonDataFile,
		os.O_RDWR|os.O_CREATE,
		0644)
	defer jsonFile.Close()
	if err != nil {
		return err
	}
	jsonFile.Write(metaData)
	return nil
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
	// lock
	filesMgr.lock.RLock()
	// unlock
	defer filesMgr.lock.RUnlock()

	// 创建日志记录器
	log := logger.NewAgent()
	defer log.EndLog()

	uvdtRootPath := path.Join(setting.AppSetting.GetRootPath(), ".uvdt")

	// 1. 当 root path 不存在时创建目录
	if _, err := os.Stat(uvdtRootPath); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create uvdt root path, %s", uvdtRootPath))
		os.MkdirAll(uvdtRootPath, os.ModeDir|os.ModePerm)
	}

	// 2. 检查创建分享文件目录 {root}/share
	//    每个分享的文件在这个目录下创建独立的目录
	sharePath := path.Join(setting.AppSetting.GetRootPath(), "share")
	if _, err := os.Stat(sharePath); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create share path %s", sharePath))
		os.MkdirAll(sharePath, os.ModeDir|os.ModePerm)
	}

	// 3. 检查文件元数据，没有则创建
	uvdtJsonDataFile := path.Join(uvdtRootPath, "uvdt.dat")
	if _, err := os.Stat(uvdtJsonDataFile); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create uvdt json data, %s", uvdtJsonDataFile))
		// blob := `{"version": "v1.0", "fileslist": [{"filename": "test.txt", "md5": "xxx"}]}`
		// blob := `{"version": "v1.0", "fileslist": []}`
		/*
			// 共享的文件
				blob := `{"version": "v1.0",
				"fileslist": [{"filename": "test.txt", "path":"share", "md5": "xxx"}]}`
		*/
		blob := `{"version": "v1.0",
		"fileslist": []}`
		f, err := os.OpenFile(uvdtJsonDataFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Err(fmt.Sprintf("Create uvdt json data fail, %s", uvdtJsonDataFile))
			return err
		}
		f.Write([]byte(blob))
		f.Close()
	}

	// 4. 从 json 文件中加载共享的文件元数据
	f, err := os.Open(uvdtJsonDataFile)
	if err != nil {
		log.Err(fmt.Sprintf("Open uvdt json data fail, %s", uvdtJsonDataFile))
		return err
	}
	defer f.Close()

	uvdtData := make([]byte, 1024<<10)
	count, err := f.Read(uvdtData)
	if err != nil && err != io.EOF {
		log.Err(fmt.Sprintf("Read uvdt data fail, %s", uvdtJsonDataFile))
		return err
	}
	uvdtData = uvdtData[:count]

	// 5. 解析元数据
	jsonMeta := make(map[string]interface{})
	if err := json.Unmarshal(uvdtData, &jsonMeta); err != nil {
		log.Err(fmt.Sprintf("Parse json uvdt data fail, %s", uvdtJsonDataFile))
		return err
	}
	filesMgr.version = jsonMeta["version"].(string)

	// 6. 创建 下载/共享 的文件管理器
	filesList := jsonMeta["fileslist"].([]interface{})
	for _, v := range filesList {
		fileInfo := v.(map[string]interface{})
		fileTasksMgr := FileTasksMgr{lock: sync.RWMutex{}}
		filesMgr.fileTasksMgr = append(filesMgr.fileTasksMgr, fileTasksMgr)
		filename := fileInfo["filename"].(string)
		md5 := fileInfo["md5"].(string)
		filepath := fileInfo["path"].(string)
		log.Info(fmt.Sprintf("Task %s[%s] started, state: %s",
			filename,
			md5,
			filepath))
		if filepath == "downloads" {
			fileTasksMgr.Start(int(setting.AppSetting.GetTaskNumForFile()),
				filename,
				md5)
		}
	}

	return nil
}

func (filesMgr *FilesManager) GetStats() (map[string]interface{}, error) {
	// lock
	filesMgr.lock.RLock()
	// unlock
	defer filesMgr.lock.RUnlock()

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

	return stats, nil
}
