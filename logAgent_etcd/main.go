package main

import (
	"dxp.GoCodeFile/logAgent/conf"
	"dxp.GoCodeFile/logAgent/etcd"
	"dxp.GoCodeFile/logAgent/kafka"
	"dxp.GoCodeFile/logAgent/taillog"
	"dxp.GoCodeFile/logAgent/utils"
	"fmt"
	"gopkg.in/ini.v1"
	"sync"
	"time"
)

var (
	//cfg *conf.AppConf//这样声明是个nil

	//这样声明是个指针的内存地址
	cfg = new(conf.AppConf)
)

//logAgent程序入口
func main() {
	//0.加载配置文件
	//结构体法
	err := ini.MapTo(cfg, "./conf/config.ini")
	if err != nil {
		fmt.Println("load ini failed,err:%v\n", err)
		return
	}
	//1.初始化kafka链接
	err = kafka.Init([]string{cfg.KafkaConf.Address}, cfg.Maxsize)
	if err != nil {
		fmt.Println("init kafka failed,err:%v\n", err)
	}
	fmt.Println("init kafka success")
	//2.初始化etcd
	err = etcd.Init(cfg.EtcdConf.Address, time.Duration(cfg.EtcdConf.Timeout)*time.Second)
	if err != nil {
		fmt.Println("init etcd failed,err:%v\n", err)
	}
	fmt.Println("init etcd success")
	//3.1 从etcd中获取日志收集项箱的配置信息
	var ipStr string
	ipStr, err = utils.GetOutboundIP()
	etcdConfKey := fmt.Sprintf(cfg.EtcdConf.Key, ipStr)
	logEntryConf, err := etcd.GetConf(etcdConfKey)
	if err != nil {
		fmt.Printf("etcd.GetConf failed,err:%v\n", err)
		return
	}
	fmt.Printf("get conf from etcd success,%v\n", logEntryConf)
	for index, value := range logEntryConf {
		fmt.Printf("index:%v  value:%v\n", index, value)
	}
	//3.2收集日志发往kafka
	taillog.Init(logEntryConf)
	//3.3派一个哨兵去监视日志收集项变化，有变化通知logAgent实现热加载配置
	//因newConfChan访问了tskMgr.newConfChan,这个channel要先在taillog.Init(logEntryConf)初始化
	newConfChan := taillog.NewConfChan() //从taillog包中获取对外暴露的通道
	var wg sync.WaitGroup
	wg.Add(1)
	go etcd.WatchConf(etcdConfKey, newConfChan)
	wg.Wait()

}
