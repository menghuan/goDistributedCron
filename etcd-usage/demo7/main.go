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
		watcher clientv3.Watcher
		getResp *clientv3.GetResponse
		err error
		watchStartRevision int64
		watchRespChan <-chan clientv3.WatchResponse
		watchResp clientv3.WatchResponse
		event *clientv3.Event
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

	//go 协程模拟etcd中的kv变化
	go func() {
		for {
			kv.Put(context.TODO(), "/cron/jobs/job7", "i am job7")
			kv.Delete(context.TODO(), "/cron/jobs/job7")
			time.Sleep(1 * time.Second)
		}
	}()

	//先get到当前的值 然后监听后续的编号
	if getResp, err = kv.Get(context.TODO(), "/cron/jobs/job7"); err != nil {
		fmt.Println(err)
		return
	}

	//判断返回值的kvs长度是否为0 不为0表示kv有值
	if len(getResp.Kvs) != 0 {
		fmt.Println("当前值:", string(getResp.Kvs[0].Value))
	}

	//当前etcd集群的事务ID 单调递增的 监听下一个revision
	watchStartRevision = getResp.Header.Revision + 1

	//创建一个watcher
	watcher = clientv3.NewWatcher(client)

	//启动一个5秒后取消的定时器
	ctx, CancelFunc := context.WithCancel(context.TODO())
    time.AfterFunc(5 * time.Second, func() {
		CancelFunc()
	})

	//启动监听
	fmt.Println("从该版本向后监听:", watchStartRevision)
    watchRespChan = watcher.Watch(ctx, "/cron/jobs/job7", clientv3.WithRev(watchStartRevision))

    //处理kv变化
    for watchResp = range watchRespChan {
    	for _,event = range watchResp.Events {
			switch event.Type {
			case mvccpb.PUT:
				fmt.Println("修改为:",string(event.Kv.Value),"Revision:",event.Kv.CreateRevision, event.Kv.ModRevision)
			case mvccpb.DELETE:
				fmt.Println("删除了","Revision:",event.Kv.ModRevision)
			}
		}
	}
}
