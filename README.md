# Pan Go Sdk
该代码库为百度网盘开放平台Go语言的SDK，详细请参考官方技术文档<https://pan.baidu.com/union/document/entrance>

## 下载
### 使用Go mod
在您的项目中的`go.mod`文件内添加这行代码
```bash
require github.com/jsyzchen/pan v0.0.5
```
并在项目中引入`github.com/jsyzchen/pan`
```go
import (
    "github.com/jsyzchen/pan/auth"
    "github.com/jsyzchen/pan/file"
)
```
### 不使用 Go mod
```bash
go get -u github.com/jsyzchen/pan/file
```

## 使用示例
[参考代码](https://github.com/jsyzchen/pan/tree/main/examples)
