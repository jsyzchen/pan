package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jsyzchen/pan/conf"
	"github.com/jsyzchen/pan/utils/httpclient"
	"log"
	"net/url"
	"strconv"
)

const (
	ListUri = "/rest/2.0/xpan/file?method=list"
	MetasUri = "/rest/2.0/xpan/multimedia?method=filemetas"
	StreamingUri = "/rest/2.0/xpan/file?method=streaming"
)

type ListResponse struct {
	conf.CloudDiskResponseBase
	List []struct {
		FsID 	uint64 `json:"fs_id"`
		Path      string `json:"path"`
		ServerFileName string `json:"server_filename"`
		Size      int    `json:"size"`
		IsDir    int    `json:"isdir"`
		Category    int    `json:"category"`
		Md5       string `json:"md5"`
		DirEmpty string `json:"dir_empty"`
		Thumbs map[string]string `json:"thumbs"`
		LocalCtime     int    `json:"local_ctime"`
		LocalMtime     int    `json:"local_mtime"`
		ServerCtime     int    `json:"server_ctime"`
		ServerMtime     int    `json:"server_mtime"`
	}
}

type MetasResponse struct {
	ErrorCode int  	 `json:"errno"`
	ErrorMsg  string `json:"errmsg"`
	RequestID int
	RequestIDStr string `json:"request_id"`
	List []struct {
		FsID 	uint64 `json:"fs_id"`
		Path      string `json:"path"`
		Category    int    `json:"category"`
		FileName string `json:"filename"`
		IsDir    int    `json:"isdir"`
		Size      int    `json:"size"`
		Md5       string `json:"md5"`
		DLink string `json:"dlink"`
		Thumbs map[string]string `json:"thumbs"`
		ServerCtime     int    `json:"server_ctime"`
		ServerMtime     int    `json:"server_mtime"`
		DateTaken int `json:"date_taken"`
		Width int `json:"width"`
		Height int `json:"height"`
	}
}

type ManagerResponse struct {
	conf.CloudDiskResponseBase
	Info []struct{
		Path string
		TaskID int
		Errno int
	}
}

type File struct {
	AccessToken string
}

func NewFileClient(accessToken string) *File {
	return &File{
		AccessToken: accessToken,
	}
}

// 获取文件列表
func (f *File) List(dir string, start, limit int) (ListResponse, error) {
	ret := ListResponse{}

	v := url.Values{}
	v.Add("access_token", f.AccessToken)
	v.Add("dir", dir)
	v.Add("start", strconv.Itoa(start))
	v.Add("limit", strconv.Itoa(limit))
	query := v.Encode()

	requestUrl := conf.OpenApiDomain + ListUri + "&" + query
	resp, err := httpclient.Get(requestUrl, map[string]string{})
	if err != nil {
		log.Println("httpclient.Get failed, err:", err)
		return ret, err
	}

	if resp.StatusCode != 200 {
		return ret, errors.New(fmt.Sprintf("HttpStatusCode is not equal to 200, httpStatusCode[%d], respBody[%s]", resp.StatusCode, string(resp.Body)))
	}

	if err := json.Unmarshal(resp.Body, &ret); err != nil {
		return ret, err
	}

	if ret.ErrorCode != 0 {//错误码不为0
		return ret, errors.New(fmt.Sprintf("error_code:%d, error_msg:%s", ret.ErrorCode, ret.ErrorMsg))
	}

	return ret, nil
}

// 通过FsID获取文件信息
func (f *File) Metas(fsIDs []uint64) (MetasResponse, error) {
	ret := MetasResponse{}

	fsIDsByte, err := json.Marshal(fsIDs)
	if err != nil {
		return ret, err
	}

	v := url.Values{}
	v.Add("access_token", f.AccessToken)
	v.Add("fsids", string(fsIDsByte))
	v.Add("dlink", "1")
	v.Add("thumb", "1")
	v.Add("extra", "1")
	query := v.Encode()

	requestUrl := conf.OpenApiDomain + MetasUri + "&" + query
	resp, err := httpclient.Get(requestUrl, map[string]string{})
	if err != nil {
		log.Println("httpclient.Get failed, err:", err)
		return ret, err
	}

	if resp.StatusCode != 200 {
		return ret, errors.New(fmt.Sprintf("HttpStatusCode is not equal to 200, httpStatusCode[%d], respBody[%s]", resp.StatusCode, string(resp.Body)))
	}

	if err := json.Unmarshal(resp.Body, &ret); err != nil {
		return ret, err
	}

	if ret.ErrorCode != 0 {//错误码不为0
		return ret, errors.New(fmt.Sprintf("error_code:%d, error_msg:%s", ret.ErrorCode, ret.ErrorMsg))
	}

	ret.RequestID, err = strconv.Atoi(ret.RequestIDStr)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

// 获取音视频在线播放地址，转码类型有M3U8_AUTO_480=>视频ts、M3U8_FLV_264_480=>视频flv、M3U8_MP3_128=>音频mp3、M3U8_HLS_MP3_128=>音频ts
func (f *File) Streaming(path string, transcodingType string) (string, error) {
	ret := ""

	v := url.Values{}
	v.Add("access_token", f.AccessToken)
	v.Add("path", path)
	v.Add("type", transcodingType)
	query := v.Encode()

	requestUrl := conf.OpenApiDomain + StreamingUri + "&" + query
	resp, err := httpclient.Get(requestUrl, map[string]string{})
	if err != nil {
		log.Println("httpclient.Get failed, err:", err)
		return ret, err
	}

	if resp.StatusCode != 200 {
		return ret, errors.New(fmt.Sprintf("HttpStatusCode is not equal to 200, httpStatusCode[%d], respBody[%s]", resp.StatusCode, string(resp.Body)))
	}

	return string(resp.Body), nil
}