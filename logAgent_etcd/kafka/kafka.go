package kafka

//往kafka写日志的模块
import (
	"fmt"
	"github.com/Shopify/sarama"
	"time"
)

type LogData struct {
	topic string
	data  string
}

var (
	client      sarama.SyncProducer //声明一个全局的链接kaf卡的生产者client
	logDataChan chan *LogData
)

//初始化client
func Init(addrs []string, maxSize int) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出一个partition
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回

	// 连接kafka
	client, err = sarama.NewSyncProducer(addrs, config)
	if err != nil {
		fmt.Println("producer closed, err:", err)
		return
	}
	//初始化lodDataChan
	logDataChan = make(chan *LogData, maxSize)
	go SendToKafka()
	return err
}

//对外部暴露的一个函数，该函数值把日志数据发送到一个内部的channel中
func SendToChan(topic, data string) {
	//构建结构体数据内容
	msg := &LogData{
		topic: topic,
		data:  data,
	}
	fmt.Println(topic, data)
	//将构建的结构体发送至channel
	logDataChan <- msg
}

//真正往kafka日志的函数
func SendToKafka() {
	for {
		select {
		case ld := <-logDataChan:
			fmt.Println("get ld success!")
			// 构造一个消息
			msg := &sarama.ProducerMessage{}
			msg.Topic = ld.topic
			msg.Value = sarama.StringEncoder(ld.data) //序列化
			// 发送消息到kafka
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				fmt.Println("send msg failed, err:", err)
				return
			}
			fmt.Printf("pid:%v offset:%v\n", pid, offset)
		default:
			time.Sleep(time.Millisecond * 50)
		}
	}

}
