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
		lease clientv3.Lease
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId clientv3.LeaseID
		kv clientv3.KV
		putResp *clientv3.PutResponse
		getResp *clientv3.GetResponse
		keepResp *clientv3.LeaseKeepAliveResponse
		keepRespChan <-chan *clientv3.LeaseKeepAliveResponse
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

	//申请一个租约
	lease = clientv3.NewLease(client)

	//申请一个10秒内的租约
	if leaseGrantResp, err = lease.Grant(context.TODO(), 10); err != nil {
		fmt.Println(err)
		return
	}

	//拿到租约的id
	leaseId = leaseGrantResp.ID

	//5秒后会取消自动续租  context.TODO() 一直续租下去
	if keepRespChan, err = lease.KeepAlive(context.TODO(), leaseId); err != nil {
		fmt.Println(err)
		return
	}

	//处理续约应答的协程
	go func() {
		for {
			select {
			case keepResp = <-keepRespChan:
				//每个续租应答的请求， 如果sdk跟etcd失联了很久 后来连接成功，但心跳失效
				if keepRespChan	== nil {
					fmt.Println("租约已经失效了")
					goto END
				} else { //每秒续租一次 所以会收到一次应答
					fmt.Println("收到自动续租应答：", keepResp.ID)
				}
			}
		}
		END:
	}()

	//获取一个kv的子集
	kv = clientv3.NewKV(client)

	//put一个kv，让其与租约关联起来，从而实现10秒自动过期
	if putResp, err = kv.Put(context.TODO(), "/cron/lock/job1", "1", clientv3.WithLease(leaseId)); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("写入成功了：", putResp.Header.Revision)

	//定时查看下key过期了没有
	for {
		if getResp, err = kv.Get(context.TODO(), "/cron/lock/job1"); err != nil {
			fmt.Println(err)
			return
		}
		if getResp.Count == 0 {
			fmt.Println("过期了")
			break
		}

		fmt.Println("还没有过期：", getResp.Kvs)
		time.Sleep(2 * time.Second)
	}

}
