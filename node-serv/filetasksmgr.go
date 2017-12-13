/*
	file tasks manager 管理
*/

package nodeserv

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/blueskyz/uvdt/logger"
)

/*
 * 上传，下载文件管理
 */

type BlockMeta struct {
	block_md5  string // 每个分片的 md5
	block_stat uint   // 0: 已完成, 1: 下载中，2: 未下载
}

type FileMeta struct {
	filename    string
	filesize    int
	block_count int // 块数量
	block_size  int // 每块大小
	blocks      []BlockMeta
}

type JobData struct {
	length uint
	data   []byte
}

// worker 定义执行具体的下载工作
type Worker struct {
	Id       int
	stop     chan bool // 退出标志
	filename string    // 文件绝对路径

	JobQueue chan JobData
}

func (w *Worker) Run() {
	go func() {
		log := logger.NewAgent()
		defer log.EndLog()

		for {
			select {
			case jobData := <-w.JobQueue:
				time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
				log.Info(fmt.Sprintf("Worker[%d] do length", jobData.length))
			case _ = <-w.stop:
				log.Info(fmt.Sprintf("Worker[%d] stop", w.Id))
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.stop <- true
}

// 文件任务管理
// 1. 管理下载的任务
// 2. 设置任务状态
type FileTasksMgr struct {
	maxDownloadThrNum int // 最大下载协程

	fileMeta     FileMeta
	downloadWkrs []*Worker
}

func (ftMgr *FileTasksMgr) Start(maxDlThrNum int, filename string) {
	ftMgr.maxDownloadThrNum = maxDlThrNum
	ftMgr.fileMeta.filename = filename
	jobQueue := make(chan JobData)
	for i := 0; i < ftMgr.maxDownloadThrNum; i++ {
		ftMgr.downloadWkrs = append(ftMgr.downloadWkrs,
			&Worker{i, make(chan bool), filename, jobQueue})
	}
}
