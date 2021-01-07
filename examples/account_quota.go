package main

import (
	"fmt"
	"github.com/jsyzchen/pan/account"
)

func main() {
	accessToken := "121.a2b8e9decca9d322acc34d7baeb3404f.YgQAQF94HC4F0g9xGgODfm0VmZ_kbKhMYOTEwHT.FiBjnQ"
	accountClient := account.NewAccountClient(accessToken)
	res, err := accountClient.Quota()
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println(res)
}
