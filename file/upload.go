package file

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/jsyzchen/pan/account"
	"github.com/jsyzchen/pan/conf"
	fileUtil "github.com/jsyzchen/pan/utils/file"
	"github.com/jsyzchen/pan/utils/httpclient"
	"github.com/syyongx/php2go"
	"io"
	"log"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type UploadResponse struct {
	conf.CloudDiskResponseBase
	Path      string `json:"path"`
	Size      int    `json:"size"`
	Ctime     int    `json:"ctime"`
	Mtime     int    `json:"mtime"`
	Md5       string `json:"md5"`
	FsID      uint64 `json:"fs_id"`
	IsDir    int    `json:"isdir"`
}

type PreCreateResponse struct {
	conf.CloudDiskResponseBase
	UploadID string `json:"uploadid"`
	Path string `json:"path"`
	ReturnType int `json:"return_type"`
	BlockList []int `json:"block_list"`
	Info UploadResponse `json:"info"`
}

type SuperFile2UploadResponse struct {
	conf.PcsResponseBase
	Md5 string `json:"md5"`
	UploadID string `json:"uploadid"`
	PartSeq string `json:"partseq"`//pcsapi PHP版本返回的是int类型，Go版本返回的是string类型
}

type LocalFileInfo struct {
	Md5 string
	Size int64
}

type Uploader struct {
	AccessToken string
	Path string
	LocalFilePath string
}

const (
	PreCreateUri = "/rest/2.0/xpan/file?method=precreate"
	CreateUri = "/rest/2.0/xpan/file?method=create"
	Superfile2UploadUri = "/rest/2.0/pcs/superfile2?method=upload"
)

func NewUploader(accessToken, path, localFilePath string) *Uploader {
	return &Uploader{
		AccessToken: accessToken,
		Path: handleSpecialChar(path),// 处理特殊字符
		LocalFilePath: localFilePath,
	}
}

// 上传文件到网盘，包括预创建、分片上传、创建3个步骤
func (u *Uploader) Upload() (UploadResponse, error) {
	var ret UploadResponse

	//1. file precreate
	preCreateRes, err := u.PreCreate()
	if err != nil {
		log.Println("PreCreate failed, err:", err)
		ret.ErrorCode = preCreateRes.ErrorCode
		ret.ErrorMsg = preCreateRes.ErrorMsg
		ret.RequestID = preCreateRes.RequestID
		return ret, err
	}

	if preCreateRes.ReturnType == 2 {//云端已存在相同文件，直接上传成功，无需请求后面的分片上传和创建文件接口
		preCreateRes.Info.ErrorCode = preCreateRes.ErrorCode
		preCreateRes.Info.ErrorMsg = preCreateRes.ErrorMsg
		preCreateRes.Info.RequestID = preCreateRes.RequestID
		return preCreateRes.Info, nil
	}

	uploadID := preCreateRes.UploadID

	//2. superfile2 upload
	fileInfo, err := u.getFileInfo()
	if err != nil {
		log.Println("getFileInfo failed, err:", err)
		return ret, err
	}
	fileSize := fileInfo.Size

	sliceSize, err := u.getSliceSize(fileSize)
	if err != nil {
		log.Println("getSliceSize failed, err:", err)
		return ret, err
	}

	sliceNum := int(math.Ceil(float64(fileSize) / float64(sliceSize)))

	//TODO 断点续传
	file, err := os.Open(u.LocalFilePath)
	if err != nil {
		return ret, err
	}
	defer file.Close()
	uploadRespChan := make(chan SuperFile2UploadResponse, sliceNum)
	sem := make(chan int, 10) //限制并发数，以防大文件上传导致占用服务器大量内存
	for i := 0; i < sliceNum; i++ {
		buffer := make([]byte, sliceSize)
		n, err := file.Read(buffer[:])
		if err != nil && err != io.EOF {
			log.Println("file.Read failed, err:", err)
			return ret, err
		}
		if n == 0 { //文件已读取结束
			break
		}

		sem <- 1 //当通道已满的时候将被阻塞
		go func(partSeq int, partByte []byte) {
			uploadResp, err := u.SuperFile2Upload(uploadID, partSeq, partByte)
			uploadRespChan <- uploadResp
			if err != nil {
				log.Printf("SuperFile2UploadFailed, partseq[%d] err[%v]", partSeq, err)
			}
			<-sem
		}(i, buffer[0:n])
	}

	blockList := make([]string, sliceNum)
	for i := 0; i < sliceNum; i++ {
		uploadResp := <-uploadRespChan
		if uploadResp.ErrorCode != 0 {//有部分文件上传失败
			log.Print("superfile2 upload part failed")
			ret.ErrorCode = uploadResp.ErrorCode
			ret.ErrorMsg = uploadResp.ErrorMsg
			ret.RequestID = uploadResp.RequestID
			return ret, errors.New("superfile2 upload part failed")
		}

		partSeq, err := strconv.Atoi(uploadResp.PartSeq)
		if err != nil {
			log.Fatalln("strconv.Atoi failed, err:", err)
			return ret, err
		}

		blockList[partSeq] = uploadResp.Md5
	}

	//3. file create
	superFile2CommitRes, err := u.Create(uploadID, blockList)
	if err != nil {
		log.Println("SuperFile2Commit failed, err:", err)
		return superFile2CommitRes, err
	}

	return superFile2CommitRes, err
}

// preCreate
func (u *Uploader) PreCreate() (PreCreateResponse, error) {
	ret := PreCreateResponse{}

	fileInfo, err := u.getFileInfo()
	if err != nil {
		log.Println("getFileInfo failed, err:", err)
		return ret, err
	}
	fileSize := fileInfo.Size
	fileMd5 := fileInfo.Md5

	sliceMd5, err := u.getSliceMd5()
	if err != nil {
		log.Println("getSliceMd5 failed, err:", err)
		return ret, err
	}

	blockList, err := u.getBlockList()
	if err != nil {
		log.Println("getBlockList failed, err:", err)
		return ret, err
	}
	blockListByte, err := json.Marshal(blockList)
	if err != nil {
		return ret, err
	}
	blockListStr := string(blockListByte)

	// path urlencode
	v := url.Values{}
	v.Add("path", u.Path)
	v.Add("size", strconv.FormatInt(fileSize, 10))
	v.Add("isdir", "0")
	v.Add("autoinit", "1")// 固定值1
	v.Add("rtype", "1")// 1 为只要path冲突即重命名
	v.Add("block_list", blockListStr)
	v.Add("content-md5", fileMd5)
	v.Add("slice-md5", sliceMd5)
	body := v.Encode()

	requestUrl := conf.OpenApiDomain + PreCreateUri + "&access_token=" + u.AccessToken
	headers := make(map[string]string)
	resp, err := httpclient.Post(requestUrl, headers, body)
	if err != nil {
		log.Println("httpclient.Post failed, err:", err)
		return ret, err
	}

	respBody := resp.Body
	if js, err := simplejson.NewJson(respBody); err == nil {
		if info, isExist := js.CheckGet("info"); isExist {//秒传返回的request_id有可能是科学计数法，这里将它统一转成uint64
			//{"return_type":2,"errno":0,"info":{"size":16877488,"category":4,"fs_id":714504460793248,"request_id":1.821160071156e+17,"path":"\/apps\/\u4e66\u68af\/easy_20210726_163824.pptx","isdir":0,"mtime":1627288705,"ctime":1627288705,"md5":"44090321ds594263c8818d7c398e5017"},"request_id":182116007115598010}
			info.Set("request_id", uint64(info.Get("request_id").MustFloat64()))
			if respBody, err = js.Encode(); err != nil {
				log.Println("simplejson Encode failed, err:", err)
				return ret, err
			}
		}
	}

	if err := json.Unmarshal(respBody, &ret); err != nil {
		log.Println("json.Unmarshal failed, err:", err)
		return ret, err
	}

	if ret.ErrorCode != 0 {//错误码不为0
		return ret, errors.New(fmt.Sprintf("error_code:%d, error_msg:%s", ret.ErrorCode, ret.ErrorMsg))
	}

	return ret, nil
}

//superfile2 upload
func (u *Uploader) SuperFile2Upload(uploadID string, partSeq int, partByte []byte) (SuperFile2UploadResponse, error) {
	ret := SuperFile2UploadResponse{}

	path := u.Path
	localFilePath := u.LocalFilePath

	// path urlencode
	v := url.Values{}
	v.Add("access_token", u.AccessToken)
	v.Add("path", path)
	v.Add("type", "tmpfile")
	v.Add("uploadid", uploadID)
	v.Add("partseq", strconv.Itoa(partSeq))
	queryParams := v.Encode()

	uploadUrl := conf.PcsDataDomain + Superfile2UploadUri + "&" + queryParams

	fileUploader := fileUtil.NewFileUploader(uploadUrl, localFilePath)
	resp, err := fileUploader.UploadByByte(partByte)
	if err != nil {
		log.Print("fileUploader.UploadByByte failed")
		return ret, err
	}

	if err := json.Unmarshal(resp, &ret); err != nil {
		log.Printf("upload failed, path[%s] response[%s]", path, string(resp))
		return ret, err
	}

	if ret.ErrorCode != 0 {//错误码不为0
		log.Printf("upload failed, path[%s] response[%s]", path, string(resp))
		return ret, errors.New(fmt.Sprintf("error_code:%d, error_msg:%s", ret.ErrorCode, ret.ErrorMsg))
	}

	return ret, nil
}

// file create
func (u *Uploader) Create(uploadID string, blockList []string) (UploadResponse, error){
	ret := UploadResponse{}

	fileInfo, err := u.getFileInfo()
	if err != nil {
		log.Println("getFileInfo failed, err:", err)
		return ret, err
	}

	blockListByte, err := json.Marshal(blockList)
	if err != nil {
		return ret, err
	}
	blockListStr := string(blockListByte)

	// path urlencode
	v := url.Values{}
	v.Add("path", u.Path)
	v.Add("uploadid", uploadID)
	v.Add("block_list", blockListStr)
	v.Add("size", strconv.FormatInt(fileInfo.Size, 10))
	v.Add("isdir", "0")
	v.Add("rtype", "1")//1 为只要path冲突即重命名
	body := v.Encode()

	requestUrl := conf.OpenApiDomain + CreateUri + "&access_token=" + u.AccessToken

	headers := make(map[string]string)
	resp, err := httpclient.Post(requestUrl, headers, body)
	if err != nil {
		log.Println("httpclient.Post failed, err:", err)
		return ret, err
	}

	if err := json.Unmarshal(resp.Body, &ret); err != nil {
		log.Printf("json.Unmarshal failed, resp[%s], err[%v]", string(resp.Body), err)
		return ret, err
	}

	if ret.ErrorCode != 0 {//错误码不为0
		log.Println("file create failed, resp:", string(resp.Body))
		return ret, errors.New(fmt.Sprintf("error_code:%d, error_msg:%s", ret.ErrorCode, ret.ErrorMsg))
	}

	return ret, nil
}

// 获取分片的大小
func (u *Uploader) getSliceSize(fileSize int64) (int64, error) {
	var sliceSize int64

	/*
	限制：
		普通用户单个分片大小固定为4MB（文件大小如果小于4MB，无需切片，直接上传即可），单文件总大小上限为4G。
		普通会员用户单个分片大小上限为16MB，单文件总大小上限为10G。
		超级会员用户单个分片大小上限为32MB，单文件总大小上限为20G。
	*/
	//切割文件，单个分片大小暂时先固定为4M，TODO 普通会员和超级会员单个分片可以更大，需判断用户的身份
	sliceSize = 4194304//4M
	accountClient := account.NewAccountClient(u.AccessToken)
	userInfo, err := accountClient.UserInfo()
	if err != nil {//获取失败直接用4M
		log.Println("account.UserInfo failed, err:", err)
		return sliceSize, nil
	}
	if userInfo.VipType == 1 {//普通会员
		sliceSize = 16777216//16M
	} else if userInfo.VipType == 2 {//超级会员
		sliceSize = 33554432//32M
	}

	if fileSize <= sliceSize {//无须切片
		sliceSize = fileSize
	}

	return sliceSize, nil
}

// 获取block_list
func (u *Uploader) getBlockList() ([]string, error) {
	blockList := []string{}

	filePath := u.LocalFilePath

	fileInfo, err := u.getFileInfo()
	if err != nil {
		log.Println("getFileInfo failed, err:", err)
		return blockList, err
	}
	fileSize := fileInfo.Size
	fileMd5 := fileInfo.Md5

	sliceSize, err := u.getSliceSize(fileSize)
	if err != nil {
		log.Println("getSliceSize failed, err:", err)
		return blockList, err
	}

	if sliceSize == fileSize {//只有一个分片
		blockList = append(blockList, fileMd5)
		return blockList, nil
	}

	buffer := make([]byte, sliceSize)
	file, err := os.Open(filePath)
	if err != nil {
		return blockList, err
	}
	defer file.Close()

	for {
		n, err := file.Read(buffer[:])
		if err != nil && err != io.EOF {
			log.Println("file.Read failed, err:", err)
			return blockList, err
		}
		if n == 0 {
			break
		}
		partBuffer := buffer[0:n]
		hash := md5.New()
		hash.Write(partBuffer)
		sliceMd5 := hex.EncodeToString(hash.Sum(nil))
		blockList = append(blockList, sliceMd5)
	}

	return blockList, nil
}

// 获取文件信息
func (u *Uploader) getFileInfo() (LocalFileInfo, error) {
	info := LocalFileInfo{}

	filePath := u.LocalFilePath
	file, err := os.Open(filePath)
	if err != nil {
		return info, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return info, err
	}

	info.Size= fileInfo.Size()

	fileMd5, err := php2go.Md5File(filePath)
	if err != nil {
		log.Println("php2go.Md5File failed, err:", err)
		return info, err
	}

	info.Md5 = fileMd5

	return info, nil
}

// 特殊字符处理，文件名里有特殊字符时无法上传到网盘，特殊字符有'\\', '?', '|', '"', '>', '<', ':', '*',"\t","\n","\r","\0","\x0B"
func handleSpecialChar(char string) string {
	specialChars := []string{"\\\\", "?", "|", "\"", ">", "<", ":", "*","\t","\n","\r","\\0","\\x0B"}

	newChar := char
	for _, specialChar := range specialChars {
		newChar = strings.Replace(newChar, specialChar, "", -1)
	}

	if newChar != char {
		fmt.Printf("char has handle, origin[%s] handled[%s]", char, newChar)
	}

	return newChar
}

// 获取分片的md5值
func (u *Uploader) getSliceMd5() (string, error) {
	var sliceMd5 string
	var sliceSize int64
	sliceSize = 262144//切割的块大小，固定为256KB

	filePath := u.LocalFilePath
	fileInfo, err := u.getFileInfo()
	if err != nil {
		log.Println("getFileInfo failed, err:", err)
		return sliceMd5, err
	}

	fileSize := fileInfo.Size
	fileMd5 := fileInfo.Md5

	if fileSize <= sliceSize {
		sliceMd5 = fileMd5
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			return sliceMd5, err
		}
		defer file.Close()

		partBuffer := make([]byte, sliceSize)
		if _, err := file.Read(partBuffer); err == nil {
			hash := md5.New()
			hash.Write(partBuffer)
			sliceMd5 = hex.EncodeToString(hash.Sum(nil))
		}
	}

	return sliceMd5, nil
}