package main

import (
	"fmt"
	"github.com/jsyzchen/pan/file"
)

func main() {
	accessToken := "122.b0a9ab31cc24b429d460cd3ce1f1af97.Yn53jGAwd_1elGgODFvYl1sp9qOYVUDRiVawin5.tbNcEw"
	fileClient := file.NewFileClient(accessToken)
	fsIDs := []uint64{765773701501523}
	res, err := fileClient.Metas(fsIDs)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println(res)
}
