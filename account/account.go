package account

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jsyzchen/pan/conf"
	"github.com/jsyzchen/pan/utils/httpclient"
	"log"
	"net/url"
)

type UserInfoResponse struct {
	BaiduName string `json:"baidu_name"`
	NetdiskName string `json:"netdisk_name"`
	AvatarUrl string `json:"avatar_url"`
	VipType int `json:"vip_type"`
	Uk int `json:"uk"` //uk字段对应auth.UserInfo方法返回的user_id
	ErrorCode int  	 `json:"errno"`
	ErrorMsg  string `json:"errmsg"`
	RequestID string `json:"request_id"`
}

type QuotaResponse struct {
	conf.CloudDiskResponseBase
	Total int `json:"total"`
	Used int `json:"used"`
	Free int `json:"free"`
	Expire bool `json:"expire"`
}

type Account struct {
	AccessToken string
}

const UserInfoUri = "/rest/2.0/xpan/nas?method=uinfo"
const QuotaUri = "/api/quota"

func NewAccountClient(accessToken string) *Account {
	return &Account{
		AccessToken: accessToken,
	}
}

// 获取网盘用户信息
func (a *Account) UserInfo() (UserInfoResponse, error) {
	ret := UserInfoResponse{}

	v := url.Values{}
	v.Add("access_token", a.AccessToken)
	query := v.Encode()

	requestUrl := conf.OpenApiDomain + UserInfoUri + "&" + query
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

// 获取用户网盘容量信息
func (a *Account) Quota() (QuotaResponse, error) {
	ret := QuotaResponse{}

	v := url.Values{}
	v.Add("access_token", a.AccessToken)
	v.Add("checkfree", "1")
	v.Add("checkexpire", "1")
	query := v.Encode()

	requestUrl := conf.OpenApiDomain + QuotaUri + "?" + query
	resp, err := httpclient.Get(requestUrl, map[string]string{})
	if err != nil {
		log.Println("httpclient.Get failed, err:", err)
		return ret, err
	}

	fmt.Println(string(resp.Body))

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
