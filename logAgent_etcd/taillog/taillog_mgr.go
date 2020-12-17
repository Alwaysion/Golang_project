package taillog

import (
	"dxp.GoCodeFile/logAgent/etcd"
	"fmt"
	"time"
)

var tskMgr *tailLogMgr

type tailLogMgr struct {
	logEntry    []*etcd.LogEntry
	tskMap      map[string]*TailTask
	newConfChan chan []*etcd.LogEntry
}

func Init(logEntryConf []*etcd.LogEntry) {
	tskMgr = &tailLogMgr{
		logEntry:    logEntryConf, //把当前日志收集项配置信息保存起来
		tskMap:      make(map[string]*TailTask, 16),
		newConfChan: make(chan []*etcd.LogEntry), //无缓冲区的通道
	}
	for _, logEntry := range logEntryConf {
		//logEntry是 *etcd.LogEntry类型
		//LogEntry.Path：要收集的日志文件的路径
		tailObj := NewTailTask(logEntry.Path, logEntry.Topic)
		//初始化的时候启动了多少个tailtask，都要记下来，为了后续判断方便
		mk := fmt.Sprintf("%s_%s", logEntry.Path, logEntry.Topic)
		tskMgr.tskMap[mk] = tailObj
	}
	go tskMgr.run()
}
func (t *tailLogMgr) run() {
	for {
		select {
		case newConf := <-t.newConfChan:
			for _, conf := range newConf {
				mk := fmt.Sprintf("%s_%s", conf.Path, conf.Topic)
				_, ok := t.tskMap[mk]
				if ok {
					//1.原来就有，不需要操作
					continue
				} else {
					//2.新增的
					tailObj := NewTailTask(conf.Path, conf.Topic)
					t.tskMap[mk] = tailObj
				}
			}
			//3.原来t.logEntry有，newConf中没有的，要删除
			for _, c1 := range t.logEntry {
				isDelete := true
				for _, c2 := range newConf {
					if c1.Path == c2.Path && c1.Topic == c2.Topic {
						isDelete = false //在新配置中找到了，不用删除
						continue
					}
				}
				fmt.Println(isDelete)
				if isDelete {
					//没找到，删除c1对应的tailObj
					mk := fmt.Sprintf("%s_%s", c1.Path, c1.Topic)
					//t.tskMap[mk]是tailObj
					t.tskMap[mk].cancelFunc()
					//delete(t.tskMap, mk)
				}
			}
			fmt.Println("new conf comming!", newConf)
		default:
			time.Sleep(time.Second)
		}
	}
}

//向外部暴露tskMgr的newConfChan(后期名GetNewConfChan)
func NewConfChan() chan<- []*etcd.LogEntry {
	return tskMgr.newConfChan
}
