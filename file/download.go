package file

import (
	"errors"
	"github.com/jsyzchen/pan/utils/file"
	"log"
)

type Downloader struct {
	LocalFilePath string
	DownloadLink string
	FsID uint64
	Path string
	AccessToken string
	TotalPart int
}

func NewDownloader(accessToken string, downloadLink string, localFilePath string, ) *Downloader {
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

//func NewDownloaderWithPath(accessToken string, path string, localFilePath string) *Downloader {
//	return &Downloader{
//		AccessToken: accessToken,
//		Path: path,
//		LocalFilePath: localFilePath,
//	}
//}

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
	} else if d.Path != "" {

	} else {
		return errors.New("param error")
	}

	if downloadLink == "" {
		return errors.New("param error, downloadLink is empty")
	}

	downloadLink += "&access_token=" + d.AccessToken
	downloader := file.NewFileDownloader(downloadLink, d.LocalFilePath)
	if err := downloader.Download(); err != nil {
		log.Println("download failed, err:", err)
		return err
	}

	return nil
}


