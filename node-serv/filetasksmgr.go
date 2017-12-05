/*
	file tasks manager 管理
*/

package nodeserv

/*
 * 上传，下载文件管理
 */

type FileMeta struct {
	filename string
}

type FileTasksMgr struct {
	maxDownloadThrNum int // 最大下载协程
	maxUploadThrNum   int // 最大上传协程
}
