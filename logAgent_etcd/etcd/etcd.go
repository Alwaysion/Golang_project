package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

var (
	cli *clientv3.Client
)

type LogEntry struct {
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

func Init(addr string, timeOut time.Duration) (err error) {
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{addr},
		DialTimeout: timeOut,
	})
	if err != nil {

	}
	fmt.Println("connect to etcd success")
	return
}

//从etcd中根据key获取配置项
func GetConf(key string) (LogEntryConf []*LogEntry, err error) {
	// get
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, key)
	cancel()
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return
	}
	for _, ev := range resp.Kvs {
		err = json.Unmarshal(ev.Value, &LogEntryConf)
		if err != nil {
			fmt.Printf("unmarshal etcd value failed,err:%v\n", err)
		}
	}
	return
}
func WatchConf(key string, newConfChan chan<- []*LogEntry) {
	ch := cli.Watch(context.Background(), key)
	//尝试从通道取值（监视的信息）
	for wresp := range ch {
		for _, evt := range wresp.Events {
			fmt.Printf("Type: %s Key:%s Value:%s\n", evt.Type, evt.Kv.Key, evt.Kv.Value)
			//通知taillog.tskMgr
			var newConf []*LogEntry
			if evt.Type != clientv3.EventTypeDelete {
				//如果不是删除操作
				err := json.Unmarshal(evt.Kv.Value, &newConf)
				if err != nil {
					fmt.Printf("unmarshal failed,err:%v\n", err)
					continue
				}
			}
			fmt.Printf("get new conf:%v\n", newConf)
			newConfChan <- newConf
		}
	}
}
