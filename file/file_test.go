package file

import (
	"github.com/jsyzchen/pan/conf"
	"testing"
)

func TestFile_List(t *testing.T) {
	fileClient := NewFileClient(conf.TestData.AccessToken)
	res, err := fileClient.List(conf.TestData.Dir, 0, 100)
	if err != nil {
		t.Errorf("TestList failed, err:%v", err)
	}
	t.Logf("TestList res: %+v", res)
}

func TestFile_Metas(t *testing.T) {
	fileClient := NewFileClient(conf.TestData.AccessToken)
	res, err := fileClient.Metas([]uint64{conf.TestData.FsID}, 0, 0)
	if err != nil {
		t.Errorf("TestMetas failed, err:%v", err)
	}
	t.Logf("TestMetas res: %+v", res)
}

func TestFile_Streaming(t *testing.T) {
	fileClient := NewFileClient(conf.TestData.AccessToken)
	res, err := fileClient.Streaming(conf.TestData.Path, conf.TestData.TranscodingType)
	if err != nil {
		t.Errorf("TestFile_Streaming failed, err:%v", err)
	}
	t.Logf("TestFile_Streaming res: %+v", res)
}

