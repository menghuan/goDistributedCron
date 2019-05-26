package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		getResp *clientv3.GetResponse
		err error
	)

	//客户端配置
	config = clientv3.Config{
		Endpoints:[]string{"47.52.153.105:2379"},
		DialTimeout: 5 * time.Second,
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	//用于读写etcd的键值对
	kv = clientv3.NewKV(client)

	if getResp,err = kv.Get(context.TODO(), "/cron/job/job1"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(getResp.Kvs, getResp.Count)
	}
}
