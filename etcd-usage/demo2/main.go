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
		putResp *clientv3.PutResponse
		err error
	)

	//客户端配置
	config = clientv3.Config{
		Endpoints:[]string{"47.52.153.105:2379"},
		DialTimeout:5 * time.Second,
	}

	//建立一个客户端连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	//创建一个kv 用于读写etcd的键值对
	kv = clientv3.NewKV(client)

	if putResp, err = kv.Put(context.TODO(), "/cron/job/job1", "bye", clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Revision", putResp.Header.Revision)
		if putResp.PrevKv != nil {
			fmt.Println("PrevValue", putResp.PrevKv.Value)
		}
	}



}
