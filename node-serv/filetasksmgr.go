/*
	file tasks manager 管理
*/

package nodeserv

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	// "sync"
	"io"
	"time"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/node-serv/setting"
)

/*
 * 上传，下载文件管理
 */

type BlockMeta struct {
	blockMd5  string // 每个分片的 md5
	blockStat uint   // 0: 已完成, 1: 下载中，2: 未下载
}

type FileMeta struct {
	filename     string
	fileMd5      string
	fileMetaName string
	filesize     int
	blockCount   int // 块数量
	blockSize    int // 每块大小
	blocks       []BlockMeta
}

func (f *FileMeta) LoadFileMeta(filePath string) (bool, error) {
	return true, nil
}

type JobData struct {
	length uint
	data   []byte
}

// worker 定义执行具体的下载工作
type Worker struct {
	id       int
	filename string // 文件绝对路径

	stop     chan bool    // 退出标志
	jobQueue chan JobData // 下载 job

	stat                  uint      // 0: 运行中，1: 下载中，2: 已停止
	lastDownloadBeginTime time.Time // 最后下载开始时间, 每完开始一次下载更新一次，用来控制下载阻塞，未完成状态的清理
	totalDownload         int64     // 总共下载的数据量，单位字节
	totalDownloadCost     int64     // 总共下载使用的时间
	errorCount            int       // 下载出错的数量
}

func (w *Worker) Run() {
	go func() {
		log := logger.NewAgent()
		defer log.EndLog()

		// 每个协程单独打开文件
		/*
			f, err := os.OpenFile(w.filename, os.O_RDWR)
			if err != nil {
				return
			}
			defer f.Close()
		*/

		for {
			select {
			case jobData := <-w.jobQueue: // 等待获取下载数据片段的任务
				// 下载数据
				w.lastDownloadBeginTime = time.Now()
				time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
				log.Info(fmt.Sprintf("Worker[%d] do length", jobData.length))

				// 写入文件
			case _ = <-w.stop: // 停止工作
				log.Info(fmt.Sprintf("Worker[%d] stop", w.id))
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.stop <- true
}

func (w *Worker) Download(jobData JobData) error {
	return nil
}

// 文件任务管理
// 1. 管理下载的任务
// 2. 设置任务状态
type FileTasksMgr struct {
	maxDownloadThrNum int // 最大下载协程

	fileMeta     FileMeta
	downloadWkrs []*Worker

	stat                  uint      // 0: 无状态（不分享），1: 下载中（分享中）， 2: 已停止，3: 等待下载，4: 分享中
	lastDownloadBeginTime time.Time // 下载开始时间
	downloadCompleteTime  time.Time // 下载完成时间
	totalDownload         int64     // 总共下载的数据量，单位字节
	totalDownloadCost     int64     // 总共下载使用的时间
}

/*
 * 创建分享文件任务
 */
func (ftMgr *FileTasksMgr) CreateShareFile(filePath string) {
}

/*
 * 创建下载任务并分享文件的任务
 *
 * 1. 当元数据目录不存在时创建目录，每个下载文件一个独立目录
 * 2. 保存种子文件: root/.uvdt/{fileMd5}/{fileMd5}.tor
 * 3. 创建下载状态文件: root/.uvdt/{fileMd5}/{fileMd5}.meta
 * 4. 下载目录不存在时创建目录，每个下载文件任务具有独立的目录
 */
func (ftMgr *FileTasksMgr) CreateDownloadFile(maxDlThrNum int,
	fileMd5 string,
	destDownloadPath string,
	torrent []byte) error {

	log := logger.NewAgent()
	defer log.EndLog()

	// 1. 当共享文件的元数据目录不存在时创建目录
	//	  创建元数据目录，每个下载文件任务具有独立的目录
	metaPath := path.Join(setting.AppSetting.GetRootPath(), ".uvdt", fileMd5)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create meta path %s", metaPath))
		os.MkdirAll(metaPath, os.ModeDir|os.ModePerm)
	}

	// 2. 保存种子文件: root/.uvdt/{fileMd5}/{fileMd5}.tor
	torFile := path.Join(metaPath, fileMd5, ".tor")
	fTorrent, err := os.OpenFile(torFile, os.O_RDWR|os.O_CREATE, 0644)
	defer fTorrent.Close()
	if err != nil {
		log.Err(fmt.Sprintf("Save torrent file fail, %s", torFile))
		return err
	}
	fTorrent.Write(torrent)

	// 3. 创建元数据目录，创建元数据文件
	//    下载状态文件: root/.uvdt/{fileMd5}/{fileMd5}.meta
	jsonMetaFile := path.Join(metaPath, fileMd5, ".meta")
	if _, err := os.Stat(jsonMetaFile); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create meta data, %s", jsonMetaFile))
		blob := `{"version": "v1.0", "fileslist": ["test.txt"]}`
		fMeta, err := os.OpenFile(jsonMetaFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Err(fmt.Sprintf("Create meta data fail, %s", jsonMetaFile))
			return err
		}
		defer fMeta.Close()
		fMeta.Write([]byte(blob))
	} else if os.IsExist(err) {
		log.Err(fmt.Sprintf("The meta data is exist, %s", jsonMetaFile))
		return err
	}

	// 4. 当下载目录不存在时创建目录
	//	  创建本地下载目录，每个下载文件任务具有独立的目录 root/downloads/downloadfile
	downloadPath := path.Join(setting.AppSetting.GetRootPath(),
		"downloads",
		destDownloadPath)
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create download path %s", downloadPath))
		os.MkdirAll(downloadPath, os.ModeDir|os.ModePerm)
	}

	// 5. 添加到本地共享文件管理的数据库

	// 6. 调用 Start 开始下载任务
	/*
		fileTasksMgr := FileTasksMgr{}
		fileTasksMgr.Start(int(setting.AppSetting.GetTaskNumForFile()), filename)
		filesMgr.fileTasksMgr = append(filesMgr.fileTasksMgr, fileTasksMgr)
	*/

	return nil
}

func (ftMgr *FileTasksMgr) Start(maxDlThrNum int, filename string) error {
	log := logger.NewAgent()
	defer log.EndLog()

	// 初始化元数据
	ftMgr.maxDownloadThrNum = maxDlThrNum

	// 1. 当下载目录不存在时创建目录
	filePath := path.Join(setting.AppSetting.GetRootPath(),
		"downloads",
		filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create download filePath, %s", filePath))
		return err
	}

	// 2. 读取元数据文件
	ftMgr.fileMeta.fileMetaName = path.Join(filePath, filename, ".meta")
	jsonMetaFile := ftMgr.fileMeta.fileMetaName

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

	// 5. 创建下载 worker
	jobQueue := make(chan JobData)
	for i := 0; i < ftMgr.maxDownloadThrNum; i++ {
		ftMgr.downloadWkrs = append(ftMgr.downloadWkrs,
			&Worker{id: i, filename: filename, stop: make(chan bool), jobQueue: jobQueue})
	}

	// 初始化统计数据

	return nil
}
