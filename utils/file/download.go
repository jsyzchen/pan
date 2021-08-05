package file

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

//FileDownloader 文件下载器
type Downloader struct {
	FileSize       int
	Link            string
	FilePath 	   string
	TotalPart      int //下载线程
	DoneFilePart   []Part
	PartSize int
}

//filePart 文件分片
type Part struct {
	Index int    //文件分片的序号
	From  int    //开始byte
	To    int    //解决byte
	Data  []byte //http下载得到的文件内容
	FilePath string //下载到本地的分片文件路径
}

//NewFileDownloader .
func NewFileDownloader(downloadLink, filePath string) *Downloader {
	return &Downloader{
		FileSize:       0,
		Link:           downloadLink,
		FilePath: 		filePath,
	}
}

func (d *Downloader) SetTotalPart(totalPart int)  {
	d.TotalPart = totalPart
}

func (d *Downloader) SetPartSize(PartSize int) {
	d.PartSize = PartSize
}

//Run 开始下载任务
func (d *Downloader) Download() error {
	if d.TotalPart == 1 {
		err := d.downloadWhole()
		return err
	}
	isSupportRange, err := d.head()
	if err != nil {
		return err
	}
	log.Println("isSupportRange:", isSupportRange)

	fileTotalSize := d.FileSize
	if d.PartSize == 0 {
		d.PartSize = 10485760 // 10M
	}

	if isSupportRange == false || fileTotalSize <= d.PartSize {//不支持Range下载或者文件比较小，直接下载文件
		err := d.downloadWhole()
		return err
	}

	log.Println("downloadPart")

	if d.TotalPart == 0 || fileTotalSize / d.PartSize < d.TotalPart {//减少range请求次数
		d.TotalPart = int(math.Ceil(float64(fileTotalSize) / float64(d.PartSize)))
	}
	maxTotalPart := 100
	if d.TotalPart > maxTotalPart {//限制分片数量
		d.TotalPart = maxTotalPart
	}

	log.Println("TotalPart:", d.TotalPart)

	d.DoneFilePart = make([]Part, d.TotalPart)
	jobs := make([]Part, d.TotalPart)
	eachSize := fileTotalSize / d.TotalPart

	for i := range jobs {
		jobs[i].Index = i
		if i == 0 {
			jobs[i].From = 0
		} else {
			jobs[i].From = jobs[i-1].To + 1
		}
		if i < d.TotalPart - 1 {
			jobs[i].To = jobs[i].From + eachSize
		} else {
			//the last filePart
			jobs[i].To = fileTotalSize - 1
		}
	}

	// 删除临时文件
	defer d.removePartFiles()

	var wg sync.WaitGroup
	isFailed := false
	sem := make(chan int, 10) //限制并发数，以防大文件下载导致占用服务器大量网络宽带和磁盘io
	for _, job := range jobs {
		wg.Add(1)
		sem <- 1 //当通道已满的时候将被阻塞
		go func(job Part) {
			defer wg.Done()
			err := d.downloadPart(job)
			if err != nil {
				log.Println("下载文件失败:", err, job)
				isFailed = true //TODO 可能会有问题
			}
			<-sem
		}(job)
	}
	wg.Wait()
	if isFailed == true {
		log.Println("下载文件失败")
		return errors.New("downloadPart failed")
	}

	return d.mergeFileParts()
}

//head 获取要下载的文件的基本信息(header) 使用HTTP Method Head
func (d *Downloader) head() (bool, error) {
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
		return isSupportRange, errors.New(fmt.Sprintf("Can't process, response is %v", resp))
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
	d.FileSize = contentLength

	return isSupportRange, nil
}

//下载分片
func (d *Downloader) downloadPart(c Part) error {
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
		if err != io.EOF && err != io.ErrUnexpectedEOF {//unexpected EOF 处理
			log.Println("ioutil.ReadAll error :", err)
			return err
		}
	}
	if len(bs) != (c.To - c.From + 1) {
		return errors.New("下载文件分片长度错误")
	}
	//c.Data = bs

	//分片文件写入到本地临时目录
	fileName := path.Base(d.FilePath)
	fileNamePrefix := fileName[0:len(path.Base(d.FilePath)) - len(path.Ext(d.FilePath))]
	nowTime := time.Now().UnixNano() / 1e6
	partFilePath := path.Join(os.TempDir(), fileNamePrefix + "_" + strconv.Itoa(c.Index) + "_" + strconv.FormatInt(nowTime, 10))
	f, err := os.Create(partFilePath)
	if err != nil {
		log.Println("open file error :", err)
		return err
	}
	// 关闭文件
	defer f.Close()
	// 字节方式写入
	_, err = f.Write(bs)
	if err != nil {
		log.Println(err)
		return err
	}
	c.FilePath = partFilePath

	d.DoneFilePart[c.Index] = c

	log.Printf("结束[%d]下载from:%d to:%d\n", c.Index, c.From, c.To)
	return nil
}

//mergeFileParts 合并下载的文件
func (d *Downloader) mergeFileParts() error {
	log.Println("开始合并文件")
	mergedFile, err := os.Create(d.FilePath)
	if err != nil {
		return err
	}
	defer mergedFile.Close()
	totalSize := 0
	for _, s := range d.DoneFilePart {
		data, err := ioutil.ReadFile(s.FilePath)
		if err != nil {
			log.Println("ioutil.ReadFile err:", err)
			return err
		}

		mergedFile.Write(data)
		totalSize += len(data)
	}
	if totalSize != d.FileSize {
		return errors.New("文件不完整")
	}
	return nil
}

// 删除临时文件
func (d *Downloader) removePartFiles() {
	var wg sync.WaitGroup
	for _, s := range d.DoneFilePart {
		if s.FilePath != "" {
			wg.Add(1)
			go func (filePath string) {
				defer wg.Done()
				if err := os.Remove(filePath); err != nil {
					log.Println(filePath, "remove failed, err:", err)
				}
			}(s.FilePath)
		}
	}
	wg.Wait()
}

//直接下载整个文件
func (d *Downloader) downloadWhole() error {
	log.Println("downloadWhole")

	// Get the data
	r, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 创建一个文件用于保存
	out, err := os.Create(d.FilePath)
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
func (d *Downloader) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(
		method,
		d.Link,
		nil,
	)
	if err != nil {
		return nil, err
	}

	r.Header.Set("User-Agent", "pan.baidu.com")
	return r, nil
}