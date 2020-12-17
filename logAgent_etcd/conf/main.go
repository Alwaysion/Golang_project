package conf

type AppConf struct {
	KafkaConf `ini:"kafka"`
	EtcdConf  `ini:"etcd"`
}
type KafkaConf struct {
	Address string `ini:"address"`
	Maxsize int    `ini:"chan_max_size"`
}
type EtcdConf struct {
	Address string `ini:"address"`
	Timeout int    `ini:"timeOut"`
	Key     string `ini:"key"`
}
