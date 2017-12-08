/*
	manager server 服务，提供管理访问
*/

package nodeserv

import (
	"github.com/blueskyz/uvdt/tracker/setting"
)

/*
 * 上传，下载文件管理
 */
type FilesManager struct {
	maxFileNum int

	fileMgr []FileTasksMgr
}

func CreateFilesMgr() FilesManager {

	// FilesManager{
}
