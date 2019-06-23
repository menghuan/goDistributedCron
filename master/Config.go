package master

import (
	"encoding/json"
	"io/ioutil"
)

//config配置文件结构
type Config struct {
	//api端口
	ApiPort          			int `json:"apiPort"`
	//api读超时
	ApiReadTimeout   			int `json:"apiReadTimeout"`
	//api写超时
	ApiWriteTimeout  			int `json:"apiWriteTimeout"`
	//etcd 集群地址
	EtcdEndpoints    			[]string `json:"etcdEndpoints"`
	//etcd超时
	EtcdDialTimeout  			int `json:"etcdDialTimeout"`
	//静态页面根目录
	StaticWebRoot    			string `json:"static_web_root"`
	//mongodburl
	MongodbUri            		string  `json:"mongodbUri"`
	//mongodb连接超时时间
	MongodbConnectTimeout 		int    `json:"mongodbConnectTimeout"`
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
