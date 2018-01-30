/*
	file tasks manager 管理
*/

package nodeserv

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"sync"
	"time"

	"github.com/blueskyz/uvdt/logger"
	"github.com/blueskyz/uvdt/node-serv/setting"
)

/*
 * 块状态
 */
const (
	BS_UNDOWNLOAD  = iota // 0
	BS_COMPLETE           // 1
	BS_DOWNLOADING        // 2
)

/*
 * 上传，下载文件管理
 */
type BlockMeta struct {
	blockMd5  string // 每个分片的 md5
	blockStat uint   // 0: 未下载，1: 已完成, 2: 下载中
	failCount uint   // 每分钟失败次数，无法下载，当大于等于10次，下1分钟内不下载此块
	lasttime  int    // 最后下载时间
}

type FileMeta struct {
	stat uint // 0: 无状态（不分享），1: 下载中（分享中）， 2: 已停止，3: 等待下载，4: 分享中

	fileMetaName string // 元文件位置
	filePath     string // 下载文件绝对路径
	fileMd5      string // 下载文件 md5
	fileSize     int    // 文件大小
	blockCount   int    // 块数量
	blockSize    int    // 每块大小
	blocks       []BlockMeta
}

func (fileMeta *FileMeta) SaveMetaFile(md5 string) error {
	log := logger.NewAgent()
	defer log.EndLog()

	// 1. 获取元数据路径
	metaPath := path.Join(setting.AppSetting.GetRootPath(), ".uvdt", md5)
	if _, err := os.Stat(metaPath); err != nil {
		log.Err(fmt.Sprintf("Get meta path %s fail", metaPath))
		return err
	}

	// 2. 获取元数据文件绝对路径
	fileMeta.fileMetaName = path.Join(metaPath, md5, ".meta")

	meta := make(map[string]interface{})
	meta["stat"] = fileMeta.stat
	meta["filepath"] = fileMeta.filePath
	meta["filemd5"] = fileMeta.fileMd5

	meta["filesize"] = fileMeta.fileSize
	if fileMeta.fileSize <= 0 {
		return errors.New(fmt.Sprintf("file size is %d", fileMeta.fileSize))
	}
	meta["blockcount"] = fileMeta.blockCount
	if fileMeta.blockCount <= 0 {
		return errors.New(fmt.Sprintf("block count is %d", fileMeta.blockCount))
	}
	meta["blocksize"] = fileMeta.blockSize
	if fileMeta.blockSize <= 0 {
		return errors.New(fmt.Sprintf("block size is %d", fileMeta.fileSize))
	}

	blocksStat := []uint8{}
	for _, v := range fileMeta.blocks {
		if v.blockStat == BS_COMPLETE {
			blocksStat = append(blocksStat, 1)
		} else {
			blocksStat = append(blocksStat, 0)
		}
	}
	meta["blocks"] = blocksStat
	metaData, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(fileMeta.fileMetaName,
		os.O_WRONLY|os.O_CREATE,
		644)
	defer f.Close()
	if err != nil {
		return err
	}
	f.Write(metaData)

	return nil
}

func (fileMeta *FileMeta) LoadMetaFile(md5 string) error {
	log := logger.NewAgent()
	defer log.EndLog()

	// 1. 获取元数据路径
	metaPath := path.Join(setting.AppSetting.GetRootPath(), ".uvdt", md5)
	if _, err := os.Stat(metaPath); err != nil {
		log.Err(fmt.Sprintf("Get meta path %s fail", metaPath))
		return err
	}

	// 2. 获取元数据文件绝对路径
	fileMeta.fileMetaName = path.Join(metaPath, md5, ".meta")
	jsonMetaFile := fileMeta.fileMetaName

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
	meta := make(map[string]interface{})
	if err := json.Unmarshal(metaData, &meta); err != nil {
		log.Err(fmt.Sprintf("Parse json meta data fail, %s", jsonMetaFile))
		return err
	}

	fileMeta.stat = meta["stat"].(uint)
	fileMeta.filePath = meta["filepath"].(string)
	fileMeta.fileMd5 = meta["filemd5"].(string)

	fileMeta.blockCount = meta["blockcount"].(int)
	fileMeta.fileSize = meta["filesize"].(int)
	fileMeta.blockSize = meta["blocksize"].(int)

	blocks := []BlockMeta{}
	for _, v := range meta["blocks"].([]interface{}) {
		if v == 1 {
			blocks = append(blocks, BlockMeta{blockStat: 1})
		} else {
			blocks = append(blocks, BlockMeta{blockStat: 0})
		}
	}
	fileMeta.blocks = blocks

	return nil
}

type JobData struct {
	pos    uint // 文件内位置
	length uint // 数据长度
}

type BlockData struct {
	workId int    // 执行下载任务的工作协程id
	pos    uint   // 文件内位置
	length uint   // 数据长度
	data   []byte // 下载的数据内容
}

// worker 定义执行具体的下载工作
type Worker struct {
	id       int
	filePath string // 文件绝对路径

	stop      chan bool      // 退出标志
	jobQueue  chan JobData   // 下载 job
	dataQueue chan BlockData // 返回下载的数据块

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

		for {
			select {
			case jobData := <-w.jobQueue: // 等待获取下载数据片段的任务
				// 下载数据
				w.lastDownloadBeginTime = time.Now()
				time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
				log.Info(fmt.Sprintf("Worker[%d] do length %d", w.id, jobData.length))
				w.dataQueue <- BlockData{workId: w.id,
					pos:    jobData.pos,
					length: jobData.length,
					data:   []byte{}}

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

// ==========================================================================
// 文件任务管理
// 1. 管理下载的任务
// 2. 设置任务状态
type FileTasksMgr struct {
	lock      sync.RWMutex
	dataQueue chan BlockData
	stop      chan bool

	maxDownloadThrNum int // 最大下载协程

	fileMeta     FileMeta
	downloadWkrs []*Worker

	/*
		0: 无状态（不分享）
		1: 下载中（分享中）
		2: 已停止
		3: 等待下载
		4: 分享中
	*/
	stat                  uint
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

func (ftMgr *FileTasksMgr) Start(maxDlThrNum int,
	filePath string,
	md5 string) error {

	ftMgr.lock.Lock()
	defer ftMgr.lock.Unlock()

	log := logger.NewAgent()
	defer log.EndLog()

	// 初始化元数据
	ftMgr.maxDownloadThrNum = maxDlThrNum

	// 1. 当下载目录不存在时创建目录
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Err(fmt.Sprintf("Create download filePath, %s", filePath))
		return err
	}

	// 2. 创建下载 worker
	jobQueue := make(chan JobData, ftMgr.maxDownloadThrNum)
	ftMgr.dataQueue = make(chan BlockData, ftMgr.maxDownloadThrNum)
	for i := 0; i < ftMgr.maxDownloadThrNum; i++ {
		ftMgr.downloadWkrs = append(ftMgr.downloadWkrs,
			&Worker{
				id:        i,
				filePath:  filePath,
				stop:      make(chan bool),
				jobQueue:  jobQueue,
				dataQueue: ftMgr.dataQueue})
	}

	for _, v := range ftMgr.downloadWkrs {
		v.Run()
	}

	//

	// 初始化统计数据

	// 创建保存数据的控制协程
	ftMgr.stop = make(chan bool)
	go func() {
		log := logger.NewAgent()
		defer log.EndLog()

		for {
			select {
			case <-time.After(time.Second): // 超时, 判断是否需要添加下载数据任务队列中
				jobQueue <- JobData{pos: 3, length: 1024}

			case blockData := <-ftMgr.dataQueue: // 等待获取下载数据片段的任务
				// 下载数据
				time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
				log.Info(fmt.Sprintf("Worker[%d] do length %d", blockData.workId, blockData.length))
				jobQueue <- JobData{pos: 3, length: 1024}

				// 写入文件
			case _ = <-ftMgr.stop: // 停止工作
				log.Info(fmt.Sprintf("Task stop"))
				return
			}
		}
	}()

	return nil
}
