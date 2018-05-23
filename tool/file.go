package tool

import (
	"os"
	"fmt"
	"io"
)

func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func Write(content string) {
	var filename = "/Users/lixin/go/src/Spider/" + GetDateTime()

	var f *os.File
	var err error

	if CheckFileIsExist(filename) {
		//如果文件存在
		f, err = os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
	} else {
		//创建文件
		f, err = os.Create(filename)
	}

	if err != nil {
		fmt.Printf("file error(%s)", err)
	}

	//写入文件(字符串)
	_, err = io.WriteString(f, content)

	if err != nil {
		fmt.Printf("wirte file error(%s)", err)
	}
}
