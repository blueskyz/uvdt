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
	BS_UNCOMPLETE         // 3: 下载未完成（下载失败）
)

/*
 * 上传，下载文件管理
 */
type BlockMeta struct {
	blockMd5  string // 每个分片的 md5
	blockStat uint   // 0: 未下载，1: 已完成, 2: 下载中，3: 下载失败
	failCount uint   // 每分钟失败次数，无法下载，当大于等于10次，下1分钟内不下载此块
	lasttime  int    // 最后下载时间
}

/*
 * 0: noshare无状态（不分享）
 * 1: download: 下载中（分享中）
 * 2: stop: 已停止
 * 3: pause: 等待下载
 * 4: share: 分享中
 */
const (
	FM_NOSHARE = iota
	FM_DOWNLOAD
	FM_STOP
	FM_PAUSE
	FM_SHARE
)

type FileMeta struct {
	version     string
	contenttype string // singlefile, multifile

	maxDlThrNum int // max thread number

	stat uint

	fileMetaName string // 元文件位置
	fileDlPath   string // 下载/共享文件绝对路径
	filename     string // 文件名
	fileMd5      string // 下载文件 md5
	fileSize     int    // 文件大小
	blockCount   int    // 块数量
	blockSize    int    // 每块大小
	blocks       []BlockMeta
}

/*
 * 上一层调用方加锁，控制并发访问的冲突
 */
func (fileMeta *FileMeta) SaveMetaFile(md5 string) error {
	log := logger.NewAgent()
	defer log.EndLog()

	// 1. 获取元数据路径
	metaPath := path.Join(setting.AppSetting.GetRootPath(), ".uvdt", md5)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create meta path %s", metaPath))
		os.MkdirAll(metaPath, os.ModeDir|os.ModePerm)
	}

	// 2. 获取元数据文件绝对路径
	fileMeta.fileMetaName = path.Join(metaPath, md5) + ".meta"
	log.Info(fmt.Sprintf("Get meta file %s", fileMeta.fileMetaName))

	meta := make(map[string]interface{})
	meta["version"] = fileMeta.version
	meta["content_type"] = fileMeta.contenttype
	meta["max_dl_thr_num"] = fileMeta.maxDlThrNum
	meta["stat"] = fileMeta.stat
	meta["file_dl_path"] = fileMeta.fileDlPath
	meta["file_name"] = fileMeta.filename
	meta["file_md5"] = fileMeta.fileMd5

	meta["file_size"] = fileMeta.fileSize
	if fileMeta.fileSize <= 0 {
		return errors.New(fmt.Sprintf("file size is %d", fileMeta.fileSize))
	}
	meta["block_count"] = fileMeta.blockCount
	if fileMeta.blockCount <= 0 {
		return errors.New(fmt.Sprintf("block count is %d", fileMeta.blockCount))
	}
	meta["block_size"] = fileMeta.blockSize
	if fileMeta.blockSize <= 0 {
		return errors.New(fmt.Sprintf("block size is %d", fileMeta.fileSize))
	}

	blocksStat := []map[string]interface{}{}
	for _, v := range fileMeta.blocks {
		block := make(map[string]interface{})
		block["md5"] = v.blockMd5
		if v.blockStat == BS_COMPLETE {
			block["bk"] = BS_COMPLETE
		} else {
			block["bk"] = BS_UNDOWNLOAD
		}
		blocksStat = append(blocksStat, block)
	}
	meta["blocks"] = blocksStat
	metaData, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(fileMeta.fileMetaName, os.O_RDWR|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	f.Write(metaData)

	return nil
}

/*
 * 上一层调用方加锁，控制并发访问的冲突
 */
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
	fileMeta.fileMetaName = path.Join(metaPath, md5) + ".meta"
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

	fileMeta.version = meta["version"].(string)
	fileMeta.contenttype = meta["content_type"].(string)
	fileMeta.maxDlThrNum = int(meta["max_dl_thr_num"].(float64))

	fileMeta.stat = uint(meta["stat"].(float64))
	fileMeta.fileDlPath = meta["file_dl_path"].(string)
	fileMeta.filename = meta["file_name"].(string)
	fileMeta.fileMd5 = meta["file_md5"].(string)

	fileMeta.fileSize = int(meta["file_size"].(float64))
	fileMeta.blockCount = int(meta["block_count"].(float64))
	fileMeta.blockSize = int(meta["block_size"].(float64))

	blocks := []BlockMeta{}
	for _, v := range meta["blocks"].([]interface{}) {
		blockMeta := BlockMeta{}
		block := v.(map[string]interface{})
		blockMeta.blockMd5 = block["md5"].(string)
		if int(block["bk"].(float64)) == 1 {
			blockMeta.blockStat = BS_COMPLETE
		} else {
			blockMeta.blockStat = BS_UNDOWNLOAD
		}
		blocks = append(blocks, blockMeta)
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
	log := logger.NewAgent()
	defer log.EndLog()

	log.Info("start download goroutine ...")

	for {
		select {
		case jobData := <-w.jobQueue: // 等待获取下载数据片段的任务
			// 下载数据
			w.lastDownloadBeginTime = time.Now()
			time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
			log.Info(fmt.Sprintf("Worker[%d] do length %d", w.id, jobData.length))
			w.Download(jobData)

			// 写入存储数据的管道
			w.dataQueue <- BlockData{workId: w.id,
				pos:    jobData.pos,
				length: jobData.length,
				data:   []byte{}}

		case _ = <-w.stop: // 停止工作
			log.Info(fmt.Sprintf("Worker[%d] stop", w.id))
			return
		}
		log.EndLog()
	}
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
func (ftMgr *FileTasksMgr) CreateShareFile(torrent []byte) (string,
	string,
	error) {

	log := logger.NewAgent()
	defer log.EndLog()

	// 1. 读取种子字符串
	torrContent := make(map[string]interface{})
	if err := json.Unmarshal(torrent, &torrContent); err != nil {
		log.Err("Parse json torrent data fail, md5.")
		return "", "", err
	}

	fileMd5 := torrContent["file_md5"].(string)
	sharePath := torrContent["file_path"].(string)

	// 2. 当共享文件的元数据目录不存在时创建目录
	//	  创建元数据目录，每个下载文件任务具有独立的目录
	metaPath := path.Join(setting.AppSetting.GetRootPath(), ".uvdt", fileMd5)
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create meta path %s", metaPath))
		os.MkdirAll(metaPath, os.ModeDir|os.ModePerm)
	}

	// 3. 保存种子文件: root/.uvdt/{fileMd5}/{fileMd5}.tor
	torFile := path.Join(metaPath, fileMd5) + ".tor"
	fTorrent, err := os.OpenFile(torFile, os.O_RDWR|os.O_CREATE, 0644)
	defer fTorrent.Close()
	if err != nil {
		log.Err(fmt.Sprintf("Save torrent file fail, %s", torFile))
		return "", "", err
	}
	fTorrent.Write(torrent)

	// 4. 当下载目录不存在时创建目录
	//	  创建本地分享目录，每个分享的文件具有独立的目录 {root}/share/{sharePath}
	abSharePath := path.Join(setting.AppSetting.GetRootPath(), sharePath)
	if _, err := os.Stat(abSharePath); os.IsNotExist(err) {
		log.Err(fmt.Sprintf("Share path not exist. %s", abSharePath))
		return "", "", err
	}
	ftMgr.fileMeta.fileDlPath = abSharePath

	ftMgr.fileMeta.version = torrContent["version"].(string)
	if ftMgr.fileMeta.version != "1.0" {
		return "", "", errors.New(fmt.Sprintf("torrent version err, %s", ftMgr.fileMeta.version))
	}

	ftMgr.fileMeta.contenttype = torrContent["contenttype"].(string)
	if ftMgr.fileMeta.contenttype != "singlefile" {
		return "", "", errors.New(fmt.Sprintf("torrent content type err, %s", ftMgr.fileMeta.contenttype))
	}

	ftMgr.fileMeta.maxDlThrNum = 0

	ftMgr.fileMeta.stat = FM_STOP
	ftMgr.fileMeta.fileDlPath = abSharePath
	ftMgr.fileMeta.filename = torrContent["file_name"].(string)
	ftMgr.fileMeta.fileMd5 = torrContent["file_md5"].(string)

	ftMgr.fileMeta.blockCount = int(torrContent["part_count"].(float64))
	ftMgr.fileMeta.fileSize = int(torrContent["file_size"].(float64))
	ftMgr.fileMeta.blockSize = int(torrContent["block_size"].(float64))

	blocks := []BlockMeta{}
	for _, v := range torrContent["file_parts"].([]interface{}) {
		blocks = append(blocks, BlockMeta{blockMd5: v.(string), blockStat: BS_UNDOWNLOAD})
	}
	ftMgr.fileMeta.blocks = blocks

	// 4. 创建元数据目录，创建元数据文件
	if err := ftMgr.fileMeta.SaveMetaFile(fileMd5); err != nil {
		log.Err(fmt.Sprintf("Save meta data fail, md5: %s", fileMd5))
		return "", "", err
	}

	//    下载状态文件: {root}/.uvdt/{fileMd5}/{fileMd5}.meta
	/*
		jsonMetaFile := path.Join(metaPath, fileMd5, ".meta")
		if _, err := os.Stat(jsonMetaFile); os.IsNotExist(err) {
			log.Info(fmt.Sprintf("Create meta data, %s", jsonMetaFile))
			blob := `{"version": "v1.0", "contenttype": singlefile, "fileslist": ["test.txt"]}`
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
	*/

	// 5. 添加到本地共享文件管理的数据库

	// 6. 调用 Start 开始下载任务
	/*
		fileTasksMgr := FileTasksMgr{}
		fileTasksMgr.Start(int(setting.AppSetting.GetTaskNumForFile()), filename)
		filesMgr.fileTasksMgr = append(filesMgr.fileTasksMgr, fileTasksMgr)
	*/

	return ftMgr.fileMeta.filename, ftMgr.fileMeta.fileMd5, nil
}

/*
 * 创建下载任务并分享文件的任务
 *
 * 1. 当元数据目录不存在时创建目录，每个下载文件一个独立目录
 * 2. 保存种子文件: {root}/.uvdt/{fileMd5}/{fileMd5}.tor
 * 3. 创建下载状态文件: {root}/.uvdt/{fileMd5}/{fileMd5}.meta
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
	torFile := path.Join(metaPath, fileMd5) + ".tor"
	fTorrent, err := os.OpenFile(torFile, os.O_RDWR|os.O_CREATE, 0644)
	defer fTorrent.Close()
	if err != nil {
		log.Err(fmt.Sprintf("Save torrent file fail, %s", torFile))
		return err
	}
	fTorrent.Write(torrent)

	// 3. 当下载目录不存在时创建目录
	//	  创建本地下载目录，每个下载文件任务具有独立的目录 {root}/downloads/{destdownloadpath}
	downloadPath := path.Join(setting.AppSetting.GetRootPath(),
		"downloads",
		destDownloadPath)
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		log.Info(fmt.Sprintf("Create download path %s", downloadPath))
		os.MkdirAll(downloadPath, os.ModeDir|os.ModePerm)
	}
	ftMgr.fileMeta.fileDlPath = downloadPath

	torrContent := make(map[string]interface{})
	if err := json.Unmarshal(torrent, &torrContent); err != nil {
		log.Err(fmt.Sprintf("Parse json torrent data fail, md5: %s", fileMd5))
		return err
	}

	ftMgr.fileMeta.version = torrContent["version"].(string)
	if ftMgr.fileMeta.version != "1.0" {
		return errors.New(fmt.Sprintf("torrent version err, %s", ftMgr.fileMeta.version))
	}

	ftMgr.fileMeta.contenttype = torrContent["contenttype"].(string)
	if ftMgr.fileMeta.contenttype != "singlefile" {
		return errors.New(fmt.Sprintf("torrent content type err, %s", ftMgr.fileMeta.contenttype))
	}

	ftMgr.fileMeta.stat = FM_STOP
	ftMgr.fileMeta.maxDlThrNum = maxDlThrNum

	ftMgr.fileMeta.fileDlPath = downloadPath
	ftMgr.fileMeta.fileMd5 = torrContent["file_md5"].(string)
	ftMgr.fileMeta.filename = torrContent["file_name"].(string)

	ftMgr.fileMeta.blockCount = int(torrContent["part_count"].(float64))
	ftMgr.fileMeta.fileSize = int(torrContent["file_size"].(float64))
	ftMgr.fileMeta.blockSize = int(torrContent["block_size"].(float64))

	blocks := []BlockMeta{}
	for _, v := range torrContent["file_parts"].([]interface{}) {
		blocks = append(blocks, BlockMeta{blockMd5: v.(string), blockStat: BS_UNDOWNLOAD})
	}
	ftMgr.fileMeta.blocks = blocks

	// 4. 创建元数据目录，创建元数据文件
	if err := ftMgr.fileMeta.SaveMetaFile(fileMd5); err != nil {
		log.Err(fmt.Sprintf("Save meta data fail, md5: %s", fileMd5))
		return err
	}

	//    下载状态文件: {root}/.uvdt/{fileMd5}/{fileMd5}.meta
	/*
		jsonMetaFile := path.Join(metaPath, fileMd5, ".meta")
		if _, err := os.Stat(jsonMetaFile); os.IsNotExist(err) {
			log.Info(fmt.Sprintf("Create meta data, %s", jsonMetaFile))
			blob := `{"version": "v1.0", "contenttype": singlefile, "fileslist": ["test.txt"]}`
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
	*/

	// 5. 添加到本地共享文件管理的数据库

	// 6. 调用 Start 开始下载任务
	/*
		fileTasksMgr := FileTasksMgr{}
		fileTasksMgr.Start(int(setting.AppSetting.GetTaskNumForFile()), filename)
		filesMgr.fileTasksMgr = append(filesMgr.fileTasksMgr, fileTasksMgr)
	*/

	return nil
}

func (ftMgr *FileTasksMgr) GetJob() JobData {
	ftMgr.lock.Lock()
	defer ftMgr.lock.Unlock()

	// 获取下一个可下载块
	return JobData{pos: 3, length: 1024}
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

	metaPath := path.Join(setting.AppSetting.GetRootPath(), ".uvdt", md5)
	jsonMetaFile := path.Join(metaPath, md5, ".meta")
	if _, err := os.Stat(jsonMetaFile); os.IsNotExist(err) {
	} else if os.IsExist(err) {
		log.Err(fmt.Sprintf("The meta data is exist, %s", jsonMetaFile))
	}
	// 加载元数据
	if err := ftMgr.fileMeta.LoadMetaFile(md5); err != nil {
		log.Err(fmt.Sprintf("Load meta data fail, md5: %s", md5))
		return err
	}

	// 1. 当下载目录不存在时创建目录
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Err(fmt.Sprintf("Create download filePath, %s", filePath))
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
		// for i, v := range ftMgr.downloadWkrs {
		// log.Info(fmt.Sprintf("download file=%s, goroutine=%d", filePath, i))
		go v.Run()
	}

	//

	// 初始化统计数据

	// 创建保存数据的控制协程
	ftMgr.stop = make(chan bool)
	go func() {
		logData := logger.NewAgent()
		defer logData.EndLog()

		logData.Info("start save data goroutine ...")

		for {
			select {
			case <-time.After(time.Second): // 超时, 判断是否需要添加下载数据任务队列中
				jobQueue <- JobData{pos: 3, length: 1024}

			case blockData := <-ftMgr.dataQueue: // 等待获取下载数据片段的任务
				// 下载数据
				time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
				logData.Info(fmt.Sprintf("save data from worker[%d], length %d",
					blockData.workId,
					blockData.length))
				jobQueue <- JobData{pos: 3, length: 1024}

				// 写入文件
			case _ = <-ftMgr.stop: // 停止工作
				logData.Info(fmt.Sprintf("Task stop"))
				return
			}
			logData.EndLog()
		}
	}()

	return nil
}
