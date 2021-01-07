package main

import (
	"fmt"
	"github.com/jsyzchen/pan/auth"
)

func main() {
	clientID := "tfk7yVXzNbTB7jnSYdfdsg"
	clientSecret := "XPOiyTivh1hnxTpiTFBqAADDfvnsql"
	redirectUri := "https://coffeephp.com"
	authClient := auth.NewAuthClient(clientID, clientSecret)
	res := authClient.OAuthUrl(redirectUri)
	fmt.Println(res)
}
