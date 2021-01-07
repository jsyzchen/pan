package auth

import (
	"github.com/jsyzchen/pan/conf"
	"testing"
)

func TestAuth_OAuthUrl(t *testing.T) {
	authClient := NewAuthClient(conf.TestData.ClientID, conf.TestData.ClientSecret)
	res := authClient.OAuthUrl(conf.TestData.RedirectUri)
	t.Logf("TestAuth_OAuthUrl res: %+v", res)
}

func TestAuth_AccessToken(t *testing.T) {
	authClient := NewAuthClient(conf.TestData.ClientID, conf.TestData.ClientSecret)
	res, err := authClient.AccessToken(conf.TestData.Code, conf.TestData.RedirectUri)
	if err != nil {
		t.Errorf("authClient.AccessToken failed, err:%v", err)
	}
	t.Logf("TestAuth_AccessToken res: %+v", res)
}

func TestAuth_RefreshToken(t *testing.T) {
	authClient := NewAuthClient(conf.TestData.ClientID, conf.TestData.ClientSecret)
	res, err := authClient.RefreshToken(conf.TestData.RefreshToken)
	if err != nil {
		t.Errorf("authClient.AccessToken failed, err:%v", err)
	}
	t.Logf("TestAuth_RefreshToken res:%+v", res)
}

func TestAuth_UserInfo(t *testing.T) {
	authClient := NewAuthClient(conf.TestData.ClientID, conf.TestData.ClientSecret)
	res, err := authClient.UserInfo(conf.TestData.AccessToken)
	if err != nil {
		t.Errorf("TestAuth_UserInfo failed, err:%v", err)
	}
	t.Logf("TestAuth_UserInfo res:%+v", res)
}
