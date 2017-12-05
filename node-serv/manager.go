/*
	manager server 服务，提供管理访问
*/

package nodeserv

/*
 * 上传，下载文件管理
 */
type FilesManager struct {
	fileMgr    []FileTasksMgr
	maxFileNum int
}
