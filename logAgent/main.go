package main

import (
	"dxp.GoCodeFile/logAgent/conf"
	"dxp.GoCodeFile/logAgent/kafka"
	"dxp.GoCodeFile/logAgent/taillog"
	"fmt"
	"gopkg.in/ini.v1"
	"time"
)

var (
	//cfg *conf.AppConf//这样声明是个nil

	//这样声明是个指针的内存地址
	cfg = new(conf.AppConf)
)

func run() {
	//1.读取日志
	for {
		select {
		case line := <-taillog.ReadChan():
			kafka.SendToKafka(cfg.KafkaConf.Topic, line.Text)
		default:
			time.Sleep(time.Second)
		}
	}

}

//logAgent程序入口
func main() {
	//0.加载配置文件
	//变量法
	//cfg, err := ini.Load("./conf/config.ini")
	//cfg_address := cfg.Section("kafka").Key("address").String()
	//cfg_topic := cfg.Section("kafka").Key("topic").String()
	//cfg_path := cfg.Section("taillog").Key("path").String()
	//结构体法
	err := ini.MapTo(cfg, "./conf/config.ini")
	if err != nil {
		fmt.Println("load ini failed,err:%v\n", err)
		return
	}
	//1.初始化kafka链接
	err = kafka.Init([]string{cfg.KafkaConf.Address})
	if err != nil {
		fmt.Println("init kafka failed,err:%v\n", err)
	}
	fmt.Println("init kafka success")
	//2.打开日志文件准备收集日志
	err = taillog.Init(cfg.TaillogConf.FileName)
	if err != nil {
		fmt.Println("init taillog failed,err:%v\n", err)
		return
	}
	fmt.Println("init taillog success")
	run()
}
