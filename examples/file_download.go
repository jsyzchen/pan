package main

import (
	"fmt"
	"github.com/jsyzchen/pan/conf"
	"github.com/jsyzchen/pan/file"
)

func main() {
	accessToken := "122.b0a9ab31cc24b429d460cd3ce1f1af97.Yn53jGAwd_1elGgODFvYl1sp9qOYVUDRiVawin5.tbNcEw"
	localFilePath := "/Download/test.jpg"

	// 方式1：通过下载地址直接下载
	dLink := "https://d.pcs.baidu.com/file/a3089c75958fb77d45b2ce6cb78fd673?fid=1426856282-250528-434991606534785&rt=pr&sign=FDtAER-DCb740ccc5511e5e8fedcff06b081203-eSDq%2FMAFhWs7qSuYaJfD3%2BbkH98%3D&expires=8h&chkbd=0&chkv=0&dp-logid=2194032036121781573&dp-callid=0&dstime=1610806466&r=446016834&origin_appid=16820976&file_type=0"
	fileDownloader := file.NewDownloader(accessToken, dLink, localFilePath)
	if err := fileDownloader.Download(); err != nil {
		fmt.Println("1.fileDownloader.Download failed, err:", err)
		return
	}
	fmt.Println("1.fileDownloader.Download success")

	// 方式2：通过文件fsID下载
	var fsID uint64
	fsID = 759719327699432
	fileDownloader = file.NewDownloaderWithFsID(accessToken, fsID, localFilePath)
	if err := fileDownloader.Download(); err != nil {
		fmt.Println("2.fileDownloader.DownloadWithFsID failed, err:", err)
		return
	}
	fmt.Println("2.fileDownloader.Download success")

	// 方式3：通过文件路径下载，非开放平台公开接口，生产环境谨慎使用
	fileDownloader = file.NewDownloaderWithPath(conf.TestData.AccessToken, conf.TestData.Path, conf.TestData.LocalFilePath)
	err := fileDownloader.Download()
	if err != nil {
		fmt.Println("3.fileDownloader.DownloaderWithPath failed, err:", err)
		return
	}
	fmt.Println("3.fileDownloader.DownloaderWithPath success")
}