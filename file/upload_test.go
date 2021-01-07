package file

import (
	"github.com/jsyzchen/pan/conf"
	"testing"
)

func TestManager_Upload(t *testing.T) {
	fileUploader := NewUploader(conf.TestData.AccessToken, conf.TestData.Path, conf.TestData.LocalFilePath)
	res, err := fileUploader.Upload()
	if err != nil {
		t.Fail()
	}
	t.Logf("TestUpload Success res: %+v", res)
}
