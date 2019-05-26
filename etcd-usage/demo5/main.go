package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		delResp *clientv3.DeleteResponse
		kvpair  *mvccpb.KeyValue
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

	//删除kv
	if delResp, err = kv.Delete(context.TODO(), "/cron/jobs/job", clientv3.WithFromKey(), clientv3.WithLimit(2)); err != nil {
		fmt.Println(err)
		return
	}
	// 被删除之前的value是什么
	if len(delResp.PrevKvs) != 0 {
		for _, kvpair = range delResp.PrevKvs {
			fmt.Println("删除了：", string(kvpair.Key), string(kvpair.Value))
		}
	}
}
