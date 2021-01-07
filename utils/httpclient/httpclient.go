package httpclient

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type HttpResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func SendRequest(method string, url string, header map[string]string, body string) (HttpResponse, error) {
	client := &http.Client{}
	var res HttpResponse
	var request *http.Request
	var err error
	if method == "POST" {
		request, err = http.NewRequest(method, url, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if method == "PUT" {
		request, err = http.NewRequest(method, url, strings.NewReader(body))
	} else {
		request, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return res, err
	}

	for k, v := range header {
		if k == "host" {
			request.Host = v
		} else {
			request.Header.Set(k, v)
		}
	}
	response, err := client.Do(request)
	if response != nil {
		res.StatusCode = response.StatusCode
		res.Header = response.Header
	}
	if err != nil {
		return res, err
	}

	defer response.Body.Close()
	res.Body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return res, err
	}
	return res, nil
}

func Post(url string, header map[string]string, body string) (HttpResponse, error) {
	return SendRequest("POST", url, header, body)
}

func Put(url string, header map[string]string, body string) (HttpResponse, error) {
	return SendRequest("PUT", url, header, body)
}

func Get(url string, header map[string]string) (HttpResponse, error) {
	return SendRequest("GET", url, header, "")
}

func Head(url string, header map[string]string) (HttpResponse, error) {
	return SendRequest("HEAD", url, header, "")
}

func Delete(url string, header map[string]string) (HttpResponse, error) {
	return SendRequest("DELETE", url, header,"")
}

// 随机获取User-Agent
func GetRandomUserAgent() string {
	userAgentList := []string{
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.1 (KHTML, like Gecko) Chrome/22.0.1207.1 Safari/537.1",
		"Mozilla/5.0 (X11; CrOS i686 2268.111.0) AppleWebKit/536.11 (KHTML, like Gecko) Chrome/20.0.1132.57 Safari/536.11",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/536.6 (KHTML, like Gecko) Chrome/20.0.1092.0 Safari/536.6",
		"Mozilla/5.0 (Windows NT 6.2) AppleWebKit/536.6 (KHTML, like Gecko) Chrome/20.0.1090.0 Safari/536.6",
		"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.1 (KHTML, like Gecko) Chrome/19.77.34.5 Safari/537.1",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/536.5 (KHTML, like Gecko) Chrome/19.0.1084.9 Safari/536.5",
		"Mozilla/5.0 (Windows NT 6.0) AppleWebKit/536.5 (KHTML, like Gecko) Chrome/19.0.1084.36 Safari/536.5",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/536.3 (KHTML, like Gecko) Chrome/19.0.1063.0 Safari/536.3",
		"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/536.3 (KHTML, like Gecko) Chrome/19.0.1063.0 Safari/536.3",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Trident/4.0; SE 2.X MetaSr 1.0; SE 2.X MetaSr 1.0; .NET CLR 2.0.50727; SE 2.X MetaSr 1.0)",
		"Mozilla/5.0 (Windows NT 6.2) AppleWebKit/536.3 (KHTML, like Gecko) Chrome/19.0.1062.0 Safari/536.3",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/536.3 (KHTML, like Gecko) Chrome/19.0.1062.0 Safari/536.3",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; 360SE)",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/536.3 (KHTML, like Gecko) Chrome/19.0.1061.1 Safari/536.3",
		"Mozilla/5.0 (Windows NT 6.1) AppleWebKit/536.3 (KHTML, like Gecko) Chrome/19.0.1061.1 Safari/536.3",
		"Mozilla/5.0 (Windows NT 6.2) AppleWebKit/536.3 (KHTML, like Gecko) Chrome/19.0.1061.0 Safari/536.3",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/535.24 (KHTML, like Gecko) Chrome/19.0.1055.1 Safari/535.24",
		"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/535.24 (KHTML, like Gecko) Chrome/19.0.1055.1 Safari/535.24",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36",
	}
	userAgentCount := len(userAgentList)

	rand.Seed(time.Now().UnixNano())
	randIndex := rand.Intn(userAgentCount)

	return userAgentList[randIndex]
}