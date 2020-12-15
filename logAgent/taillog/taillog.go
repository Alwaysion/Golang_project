package taillog

import (
	"fmt"
	"github.com/hpcloud/tail"
)

var (
	tailObj *tail.Tail
	LogChan chan string
)

//从日志文件收集日志
func Init(filename string) (err error) {
	config := tail.Config{
		ReOpen:    true,                                 // 文件到了一定大小就重新打开
		Follow:    true,                                 // 是否跟随，备份之后跟随
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件哪个位置开始读
		MustExist: false,                                // 上面的文件若不存在则报错
		Poll:      true,                                 // Poll for file changes instead of using inotify
	}
	tailObj, err = tail.TailFile(filename, config)
	if err != nil {
		fmt.Println("tail file err:", err)
		return
	}
	return
}

func ReadChan() <-chan *tail.Line {
	return tailObj.Lines
}
