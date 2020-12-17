package taillog

import (
	"context"
	"dxp.GoCodeFile/logAgent/kafka"
	"fmt"
	"github.com/hpcloud/tail"
)

//从日志文件收集日志
type TailTask struct {
	path     string
	topic    string
	instance *tail.Tail
	//为了能实现退出t.run()
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewTailTask(path, topic string) (tailObj *TailTask) {
	ctx, cancel := context.WithCancel(context.Background())
	tailObj = &TailTask{
		path:       path,
		topic:      topic,
		ctx:        ctx,
		cancelFunc: cancel,
	}
	tailObj.init() //根据路径去打开相应的日志
	return
}
func (t *TailTask) init() {
	config := tail.Config{
		ReOpen:    true,                                 // 文件到了一定大小就重新打开
		Follow:    true,                                 // 是否跟随，备份之后跟随
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件哪个位置开始读
		MustExist: false,                                // 上面的文件若不存在则报错
		Poll:      true,                                 // Poll for file changes instead of using inotify
	}
	var err error
	t.instance, err = tail.TailFile(t.path, config)
	if err != nil {
		fmt.Println("tail file err:", err)
	}
	//当goroutine执行的函数退出的时候，goroutine就结束了
	go t.run() //采集日志发送到kafka
}
func (t *TailTask) run() {
	for {
		select {
		case <-t.ctx.Done():
			fmt.Printf("tail task:%s_%s 结束了...\n", t.path, t.topic)
			return
		case line := <-t.instance.Lines:
			//先把日志发送到通道中
			kafka.SendToChan(t.topic, line.Text)
			//kafka那个包中有单独的goroutine去取日志发送到kafka
		}
	}
}
