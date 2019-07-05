package register

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"goDistributedCron/common"
	"goDistributedCron/worker"
	"net"
	"time"
)

// 注册节点到etcd： /distributed_cron/workers/IP地址
type JobRegister struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease

	localIP string // 本机IP
}

var (
	G_jobRegister *JobRegister
)

// 获取本机网卡IP
func getLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet // IP地址
		isIpNet bool
	)
	// 获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	// 取第一个非lo的网卡IP
	for _, addr = range addrs {
		// 这个网络地址是IP地址: ipv4, ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String() // 192.168.1.1
				return
			}
		}
	}

	err = common.ERR_NO_FOUND_LOCAL_IP
	return
}

// 注册到/distributed_cron/workers/IP, 并自动续租
func (jobRegister *JobRegister) workerListKeepOnline() {
	var (
		regKey         string
		leaseGrantResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)

	for {
		// 注册路径
		regKey = common.ETCD_JOB_WORKER_DIR + jobRegister.localIP

		cancelFunc = nil

		// 创建租约
		if leaseGrantResp, err = jobRegister.lease.Grant(context.TODO(), 10); err != nil {
			goto KEEPRETRY
		}

		// 自动续租
		if keepAliveChan, err = jobRegister.lease.KeepAlive(context.TODO(), leaseGrantResp.ID); err != nil {
			goto KEEPRETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		// 注册到etcd
		if _, err = jobRegister.kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			goto KEEPRETRY
		}

		// 处理续租应答
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil { // 续租失败
					goto KEEPRETRY
				}
			}
		}

	KEEPRETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}

//初始化注册发现 基于etcd
func InitJobRegister() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		localIp string
	)

	// 初始化配置
	config = clientv3.Config{
		Endpoints:   worker.G_config.EtcdEndpoints,                                     // 集群地址
		DialTimeout: time.Duration(worker.G_config.EtcdDialTimeout) * time.Millisecond, // 连接超时
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	// 本机IP
	if localIp, err = getLocalIP(); err != nil {
		return
	}

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	G_jobRegister = &JobRegister{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIp,
	}

	// 服务注册
	go G_jobRegister.workerListKeepOnline()

	return
}
