package main

import (
	"fmt"
	"github.com/jsyzchen/pan/file"
)

func main() {
	accessToken := "122.b0a9ab31cc24b429d460cd3ce1f1af97.Yn53jGAwd_1elGgODFvYl1sp9qOYVUDRiVawin5.tbNcEw"
	fileClient := file.NewFileClient(accessToken)
	path := "/apps/书梯/CHSS.mkv"
	transcodingType := "M3U8_AUTO_480"
	res, err := fileClient.Streaming(path, transcodingType)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println(res)
}
