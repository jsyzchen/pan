package file

import (
	"bytes"
	"github.com/jsyzchen/pan/utils/httpclient"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type Uploader struct {
	Url string
	FilePath string
}

//NewFileUploader
func NewFileUploader(url, filePath string) *Uploader {
	return &Uploader{
		Url: url,
		FilePath: filePath,
	}
}

// 上传文件
func (u *Uploader) Upload() ([]byte, error) {
	ret := []byte("")

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	//"file" 为接收时定义的参数名
	fileWriter, err := bodyWriter.CreateFormFile("file", filepath.Base(u.FilePath))
	if err != nil {
		log.Println("error writing to buffer, err:", err)
		return ret, err
	}

	//打开文件
	fh, err := os.Open(u.FilePath)
	if err != nil {
		log.Println("error opening file, err:", err)
		return ret, err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return ret, err
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	//提交请求
	request, err := http.NewRequest("POST", u.Url, bodyBuf)
	if err != nil {
		return ret, err
	}

	request.Header.Add("Content-Type", contentType)
	//随机设置一个User-Agent
	userAgent := httpclient.GetRandomUserAgent()
	request.Header.Set("User-Agent", userAgent)

	//处理返回结果
	client := &http.Client{}
	resp, err := client.Do(request)
	//打印接口返回信息
	if err != nil {
		log.Println("request uploadUrl failed, err:", err)
		return ret, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}
	//根据实际需要，返回相应的信息
	return respBody, nil
}

//直接通过字节上传
func (u *Uploader) UploadByByte(fileByte []byte) ([]byte, error) {
	ret := []byte("")

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)
	//"file" 为接收时定义的参数名
	fileWriter, err := bodyWriter.CreateFormFile("file", filepath.Base(u.FilePath))
	if err != nil {
		log.Println("error writing to buffer, err:", err)
		return ret, err
	}

	//iocopy
	_, err = io.Copy(fileWriter, bytes.NewReader(fileByte))
	if err != nil {
		return ret, err
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	//提交请求
	request, err := http.NewRequest("POST", u.Url, bodyBuf)
	if err != nil {
		return ret, err
	}

	request.Header.Add("Content-Type", contentType)
	//随机设置一个User-Agent
	userAgent := httpclient.GetRandomUserAgent()
	request.Header.Set("User-Agent", userAgent)

	//处理返回结果
	client := &http.Client{}
	resp, err := client.Do(request)
	//打印接口返回信息
	if err != nil {
		log.Println("上传错误信息：", err)
		return ret, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}
	//根据实际需要，返回相应的信息
	return respBody, nil
}
