//百度授权相关，wiki地址 https://openauth.baidu.com/doc/doc.html
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jsyzchen/pan/conf"
	"github.com/jsyzchen/pan/utils/httpclient"
	"log"
	"net/url"
)

type Auth struct {
	ClientID string
	ClientSecret string
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope string `json:"scope"`
	SessionKey string `json:"session_key"`
	SessionSecret string `json:"session_secret"`
	Error string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope string `json:"scope"`
	SessionKey string `json:"session_key"`
	SessionSecret string `json:"session_secret"`
	Error string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type UserInfoResponse struct {
	OpenID string `json:"openid"`
	UnionID string `json:"unionid"` // 百度用户统一标识，对当前开发者帐号唯一
	UserID string `json:"userid"` // 老版百度用户的唯一标识，后续不在返回该字段，user_id字段对应account.UserInfo方法返回的uk
	UserName string `json:"username"`
	SecureMobile int `json:"securemobile"` // 当前用户绑定手机号，需要向百度开放平台单独申请权限
	Portrait string `json:"portrait"`
	UserDetail string `json:"userdetail"`
	Birthday string `json:"birthday"`
	Marriage string `json:"marriage"`
	Sex string `json:"sex"`
	Blood string `json:"blood"`
	IsBindMobile string `json:"is_bind_mobile"`
	IsRealName string `json:"is_realname"`
	ErrorCode int  	 `json:"errno"`
	ErrorMsg  string `json:"errmsg"`
}

const OAuthUri = "/oauth/2.0/authorize"
const OAuthTokenUri = "/oauth/2.0/token"
const UserInfoUri = "/rest/2.0/passport/users/getInfo"

func NewAuthClient(clientID string, clientSecret string) *Auth {
	return &Auth{
		ClientID: clientID,
		ClientSecret: clientSecret,
	}
}

// 获取授权页网址
func (a *Auth) OAuthUrl(redirectUri string) string {
	oAuthUrl := ""

	v := url.Values{}
	v.Add("response_type", "code")
	v.Add("client_id", a.ClientID)
	v.Add("redirect_uri", redirectUri)
	v.Add("scope", "basic,netdisk")
	v.Add("state", "STATE")
	query := v.Encode()

	oAuthUrl = conf.BaiduOpenApiDomain + OAuthUri + "?" + query

	return oAuthUrl
}

// 获取AccessToken
func (a *Auth) AccessToken(code, redirectUri string) (AccessTokenResponse, error) {
	ret := AccessTokenResponse{}

	v := url.Values{}
	v.Add("grant_type", "authorization_code")
	v.Add("code", code)
	v.Add("client_id", a.ClientID)
	v.Add("client_secret", a.ClientSecret)
	v.Add("redirect_uri", redirectUri)
	query := v.Encode()

	requestUrl := conf.BaiduOpenApiDomain + OAuthTokenUri + "?" + query

	resp, err := httpclient.Get(requestUrl, map[string]string{})
	if err != nil {
		log.Println("httpclient.Get failed, err:", err)
		return ret, err
	}

	if resp.StatusCode != 200 {
		return ret, errors.New(fmt.Sprintf("HttpStatusCode is not equal to 200, httpStatusCode[%d], respBody[%s]", resp.StatusCode, resp.Body))
	}

	if err := json.Unmarshal(resp.Body, &ret); err != nil {
		return ret, err
	}

	if ret.Error != "" {//有错误
		return ret, errors.New(ret.ErrorDescription)
	}

	return ret, nil
}

// 刷新AccessToken
func (a *Auth) RefreshToken(refreshToken string) (RefreshTokenResponse, error) {
	ret := RefreshTokenResponse{}

	v := url.Values{}
	v.Add("grant_type", "refresh_token")
	v.Add("refresh_token", refreshToken)
	v.Add("client_id", a.ClientID)
	v.Add("client_secret", a.ClientSecret)
	query := v.Encode()

	requestUrl := conf.BaiduOpenApiDomain + OAuthTokenUri + "?" + query

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

	if ret.Error != "" {//有错误
		return ret, errors.New(ret.ErrorDescription)
	}

	return ret, nil
}

// 获取授权用户的百度账号信息，可以通过unionid字段来识别多个百度产品授权的是否是同一用户
// 注：获取网盘账号信息请使用account.UserInfo方法
func (a *Auth) UserInfo(accessToken string) (UserInfoResponse, error) {
	ret := UserInfoResponse{}

	v := url.Values{}
	v.Add("access_token", accessToken)
	v.Add("get_unionid", "1")//需要获取unionid时，传递get_unionid = 1
	query := v.Encode()

	requestUrl := conf.BaiduOpenApiDomain + UserInfoUri + "?" + query

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

	if ret.ErrorCode != 0 {//有错误
		return ret, errors.New(ret.ErrorMsg)
	}

	return ret, nil
}