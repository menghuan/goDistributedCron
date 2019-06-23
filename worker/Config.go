package worker

import (
	"encoding/json"
	"io/ioutil"
)

//config配置文件结构
type Config struct {
	//etcd 集群地址
	EtcdEndpoints    		[]string 	`json:"etcdEndpoints"`
	//etcd超时
	EtcdDialTimeout  		int 		`json:"etcdDialTimeout"`
	//mongodb地址
	MongodbUri       		string   	`json:"mongodbUri"`
	//mongodb链接超时时间
	MongodbConnectTimeOut  	int  		`json:"mongodbConnectTimeOut"`
	//日志批次大小
	JobLogBatchSize         int      	`json:"jobLogBatchSize"`
	//日志提交超时时间
	JobLogCommitTimeout     int      	`json:"jobLogCommitTimeout"`
}

var (
	//配置文件单例对象
	G_config *Config
)

//加载配置 通过传递文件名 进行读取 反序列化解析
func InitConfig(filename string) (err error)  {
	var (
		content []byte
		conf    Config
	)

	//1.读取配置文件
	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	//2.json反序列化
	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}

	//3. 赋值单例对象
	G_config = &conf

	return
}
