package file

import (
	"errors"
	"fmt"
	"github.com/jsyzchen/pan/utils/httpclient"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
)

//FileDownloader 文件下载器
type FileDownloader struct {
	fileSize       int
	url            string
	filePath 	   string
	totalPart      int //下载线程
	doneFilePart   []filePart
}

//filePart 文件分片
type filePart struct {
	Index int    //文件分片的序号
	From  int    //开始byte
	To    int    //解决byte
	Data  []byte //http下载得到的文件内容
}

//NewFileDownloader .
func NewFileDownloader(url, filePath string, totalPart int) *FileDownloader {
	return &FileDownloader{
		fileSize:       0,
		url:            url,
		filePath: 		filePath,
		totalPart:      totalPart,
		doneFilePart:   make([]filePart, totalPart),
	}
}

//Run 开始下载任务
func (d *FileDownloader) Download() error {
	if d.totalPart == 1 {
		err := d.downloadWholeFile()
		return err
	}
	isSupportRange, err := d.head()
	if err != nil {
		return err
	}
	fileTotalSize := d.fileSize
	smallFileSize := 2097152 // 2M
	if isSupportRange == false || fileTotalSize <= smallFileSize {//不支持Range下载或者文件比较小，直接下载文件
		err := d.downloadWholeFile()
		return err
	}

	if fileTotalSize / smallFileSize < d.totalPart {//减少range请求次数
		d.totalPart = int(math.Ceil(float64(fileTotalSize) / float64(smallFileSize)))
	}

	jobs := make([]filePart, d.totalPart)
	eachSize := fileTotalSize / d.totalPart

	for i := range jobs {
		jobs[i].Index = i
		if i == 0 {
			jobs[i].From = 0
		} else {
			jobs[i].From = jobs[i-1].To + 1
		}
		if i < d.totalPart - 1 {
			jobs[i].To = jobs[i].From + eachSize
		} else {
			//the last filePart
			jobs[i].To = fileTotalSize - 1
		}
	}

	var wg sync.WaitGroup
	for _, j := range jobs {
		wg.Add(1)
		go func(job filePart) {
			defer wg.Done()
			err := d.downloadPart(job)
			if err != nil {
				log.Println("下载文件失败:", err, job)
			}
		}(j)
	}
	wg.Wait()
	return d.mergeFileParts()
}

//head 获取要下载的文件的基本信息(header) 使用HTTP Method Head
func (d *FileDownloader) head() (bool, error) {
	isSupportRange := false
	r, err := d.getNewRequest("HEAD")
	if err != nil {
		return isSupportRange, err
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return isSupportRange, err
	}
	if resp.StatusCode > 299 {
		return isSupportRange, errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
	}
	//检查是否支持 断点续传
	if resp.Header.Get("Accept-Ranges") == "bytes" {
		isSupportRange = true
	}

	//获取文件大小
	contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return isSupportRange, err
	}
	d.fileSize = contentLength

	return isSupportRange, nil
}

//下载分片
func (d *FileDownloader) downloadPart(c filePart) error {
	r, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}
	log.Printf("开始[%d]下载from:%d to:%d\n", c.Index, c.From, c.To)
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", c.From, c.To))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("服务器错误状态码: %v", resp.StatusCode))
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(bs) != (c.To - c.From + 1) {
		return errors.New("下载文件分片长度错误")
	}
	c.Data = bs
	d.doneFilePart[c.Index] = c
	return nil
}

//mergeFileParts 合并下载的文件
func (d *FileDownloader) mergeFileParts() error {
	log.Println("开始合并文件")
	mergedFile, err := os.Create(d.filePath)
	if err != nil {
		return err
	}
	defer mergedFile.Close()
	totalSize := 0
	for _, s := range d.doneFilePart {
		mergedFile.Write(s.Data)
		totalSize += len(s.Data)
	}
	if totalSize != d.fileSize {
		return errors.New("文件不完整")
	}
	return nil
}

//直接下载整个文件
func (d *FileDownloader) downloadWholeFile() error {
	url := d.url

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 创建一个文件用于保存
	out, err := os.Create(d.filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 然后将响应流和文件流对接起来
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// getNewRequest 创建一个request
func (d *FileDownloader) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(
		method,
		d.url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	//随机设置一个User-Agent
	userAgent := httpclient.GetRandomUserAgent()
	r.Header.Set("User-Agent", userAgent)
	return r, nil
}