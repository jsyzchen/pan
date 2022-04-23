package file

import (
	"errors"
	"github.com/jsyzchen/pan/account"
	"github.com/jsyzchen/pan/conf"
	"github.com/jsyzchen/pan/utils/file"
	"log"
	"net/url"
)

type Downloader struct {
	LocalFilePath string
	DownloadLink string
	FsID uint64
	Path string
	AccessToken string
	TotalPart int
}

const (
	PcsFileDownloadUri = "/rest/2.0/pcs/file?method=download"
)

func NewDownloader(accessToken string, downloadLink string, localFilePath string) *Downloader {
	return &Downloader{
		AccessToken: accessToken,
		LocalFilePath: localFilePath,
		DownloadLink: downloadLink,
	}
}

func NewDownloaderWithFsID(accessToken string, fsID uint64, localFilePath string) *Downloader {
	return &Downloader{
		AccessToken: accessToken,
		FsID: fsID,
		LocalFilePath: localFilePath,
	}
}

// 非开放平台公开接口，生产环境谨慎使用
func NewDownloaderWithPath(accessToken string, path string, localFilePath string) *Downloader {
	return &Downloader{
		AccessToken: accessToken,
		Path: path,
		LocalFilePath: localFilePath,
	}
}

// 执行下载
func (d *Downloader) Download() error {
	downloadLink := ""
	if d.LocalFilePath == "" || d.AccessToken == "" {
		return errors.New("param error, localFilePath is empty")
	}

	if d.DownloadLink != "" {//直接下载
		downloadLink = d.DownloadLink
	} else if d.FsID != 0 {
		// 根据fsID获取下载链接
		fileClient := NewFileClient(d.AccessToken)
		metas, err := fileClient.Metas([]uint64{d.FsID})
		if err != nil {
			log.Println("fileClient.Metas failed, err:", err)
			return err
		}
		if len(metas.List) == 0 {
			log.Println("file don't exist")
			return errors.New("file don't exist")
		}
		downloadLink = metas.List[0].DLink
	} else if d.Path != "" { // TODO 如何通过文件路径获取下载地址
		v := url.Values{}
		v.Add("path", d.Path)
		v.Add("access_token", d.AccessToken)
		body := v.Encode()
		downloadLink = conf.PcsApiDomain + PcsFileDownloadUri + "&" + body
	} else {
		return errors.New("param error")
	}

	if downloadLink == "" {
		return errors.New("param error, downloadLink is empty")
	}

	downloadLink += "&access_token=" + d.AccessToken
	downloader := file.NewFileDownloader(downloadLink, d.LocalFilePath)

	accountClient := account.NewAccountClient(d.AccessToken)
	if userInfo, err := accountClient.UserInfo(); err == nil {
		log.Println("VipType:", userInfo.VipType)
		if userInfo.VipType == 2 { //当前用户是超级会员
			downloader.SetPartSize(52428800) //设置每分片下载文件大小，50M
			downloader.SetCoroutineNum(10) //分片下载并发数，普通用户不支持并发分片下载
		}
	}

	if err := downloader.Download(); err != nil {
		log.Println("download failed, err:", err)
		return err
	}

	return nil
}


