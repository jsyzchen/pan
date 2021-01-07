package main

import (
	"fmt"
	"github.com/jsyzchen/pan/auth"
)

func main() {
	clientID := "tfk7yVXzNbTB7jnSYdfdsg"
	clientSecret := "XPOiyTivh1hnxTpiTFBqAADDfvnsql"
	refreshToken := "122.b0a9ab31cc24b429d460cd3ce1f1af97.Yn53jGAwd_1elGgODFvYl1sp9qOYVUDRiVawin5.tbNcEw"
	authClient := auth.NewAuthClient(clientID, clientSecret)
	res, err := authClient.RefreshToken(refreshToken)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println(res)
}
