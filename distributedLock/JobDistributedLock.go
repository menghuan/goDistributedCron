package distributedLock

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"goDistributedCron/common"
)

//基于etcd的分布式锁 （事务TXN）
type JobDistributedLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string             //任务名
	cancleFunc context.CancelFunc //用于终止自动续租
	leaseId    clientv3.LeaseID   // 租约ID
	isLocked   bool               // 是否上锁成功
}

//初始化分布式锁对象
func InitJobDistributedLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobDistributedLock *JobDistributedLock) {
	jobDistributedLock = &JobDistributedLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
	return
}

//尝试分布式锁 基于etcd
func (jobDistributedLock *JobDistributedLock) TryJobDistributedLock() (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancleCtx      context.Context
		cancleFunc     context.CancelFunc
		leaseId        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		txn            clientv3.Txn
		lockKey        string
		txtResp        *clientv3.TxnResponse
	)
	//1.创建租约（5秒）
	if leaseGrantResp, err = jobDistributedLock.lease.Grant(context.TODO(), 5); err != nil {
		return
	}

	//context用于自动取消续租
	cancleCtx, cancleFunc = context.WithCancel(context.TODO())

	//租约id
	leaseId = leaseGrantResp.ID

	//2.自动续租
	if keepRespChan, err = jobDistributedLock.lease.KeepAlive(cancleCtx, leaseId); err != nil {
		goto LOCKFAIL
	}

	//3.处理续租应答的协程
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepResp = <-keepRespChan: //自动续租的应答
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()

	//4.创建事务TXN
	txn = jobDistributedLock.kv.Txn(context.TODO())

	//锁路径
	lockKey = common.ETCD_JOB_LOCK_DIR + jobDistributedLock.jobName

	//5.事务抢锁  比较锁版本 锁路径的创建版本 = 0 即不存在的话
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		//如果不存在就put一个空锁并带一个租约过去 这样本节点宕机也不会占用这把锁，其他
		//其他在线的worker节点可以继续调度任务
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		//如果锁被占的话 就直接获取一下这把锁
		Else(clientv3.OpGet(lockKey))

	//提交事务
	if txtResp, err = txn.Commit(); err != nil {
		//提交失败后 取消续租和释放租约 保证任何异常下都能回滚掉
		goto LOCKFAIL
	}

	//6.成功返回，失败释放租约
	if !txtResp.Succeeded { // if条件不成立 锁被占用
		err = common.ERR_DISTRIBUTED_ALREADY_LOCK
		goto LOCKFAIL
	}

	//抢锁成功
	jobDistributedLock.leaseId = leaseId
	jobDistributedLock.cancleFunc = cancleFunc
	jobDistributedLock.isLocked = true
	return

LOCKFAIL:
	//取消自动续租
	cancleFunc()
	//释放租约
	jobDistributedLock.lease.Revoke(context.TODO(), leaseId)

	return
}

//释放分布式锁
func (jobDistributedLock *JobDistributedLock) UnLock() {
	//当锁成功后
	if jobDistributedLock.isLocked {
		//取消自动续租的协程
		jobDistributedLock.cancleFunc()
		//释放租约
		jobDistributedLock.lease.Revoke(context.TODO(), jobDistributedLock.leaseId)
	}
}
