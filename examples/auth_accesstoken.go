package main

import (
	"fmt"
	"github.com/jsyzchen/pan/auth"
)

func main() {
	clientID := "tfk7yVXzNbTB7jnSYdfdsg"
	clientSecret := "XPOiyTivh1hnxTpiTFBqAADDfvnsql"
	code := "746ab1956b0b221221b3dc0c3dce7362"
	redirectUri := "https://coffeephp.com"
	authClient := auth.NewAuthClient(clientID, clientSecret)
	res, err := authClient.AccessToken(code, redirectUri)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println(res)
}
