/*
	manager server 服务，提供管理访问
*/

package nodeserv

import (
	"github.com/blueskyz/uvdt/node-serv/setting"
)

func CreateFilesMgr() *FilesManager {

	filesMgr := &FilesManager{maxFileNum: setting.AppSetting.GetMaxFileNum()}
	// FilesManager{
	return filesMgr
}

/*
 * 上传，下载文件管理
 */
type FilesManager struct {
	maxFileNum uint

	fileTasksMgr []*FileTasksMgr
}

func (filesMgr FilesManager) GetMaxFileNum() uint {
	return filesMgr.maxFileNum
}

func (filesMgr FilesManager) GetCurrentFileNum() int {
	return len(filesMgr.fileTasksMgr)
}
