package master

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"goDistributedCron/common"
	"time"
)

//任务管理器 结构体
type JobManager struct {
	client   *clientv3.Client
	kv       clientv3.KV
	lease    clientv3.Lease
}

var (
	//任务管理器单例对象
	G_jobMgr *JobManager
)

//初始化任务管理器 主要是etcd 增删改查 返回err
func InitJobManager() (err error){
	var (
		config    clientv3.Config
		client    *clientv3.Client
		kv        clientv3.KV
		lease     clientv3.Lease
	)
	//初始化etcd配置
	config = clientv3.Config{
		//集群配置
		Endpoints:G_config.EtcdEndpoints,
		//连接超时
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	//得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	//赋值单例
	G_jobMgr = &JobManager{
		client:client,
		kv:kv,
		lease:lease,
	}
	return
}

//面向对象方式 类方法SaveJob 返回之前的job信息
func(jobMgr *JobManager) SaveJob(job *common.Job) (oldJob *common.Job, err error){
	// 把任务保存到/cron/jobs/任务名 -> json
	var (
		jobKey      string
		jobValue    []byte
		putResp 	*clientv3.PutResponse
		oldJobObj 	common.Job
	)

	//etcd 保存任务的key
	jobKey = common.ETCD_JOB_SAVE_DIR + job.Name
	//任务信息json
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	//保存到etcd中 设置clientv3.WithPrevKV() 返回值获取putResp.PrevKv 之前的值
	if putResp, err = jobMgr.kv.Put(context.TODO(),jobKey, string(jobValue),clientv3.WithPrevKV()); err != nil {
		return
	}
	//如果是更新，那么返回旧的值
	if putResp.PrevKv != nil {
		//对旧值做一个反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value,&oldJobObj); err != nil {
			//不管旧值是否正确都返回正确
			err = nil
			return
		}
		//赋值旧job值
		oldJob = &oldJobObj
	}
	return
}

//删除任务
func (jobMgr *JobManager) DeleteJob(name string) (oldJob *common.Job, err error)  {
	var (
		jobKey    string
		delResp   *clientv3.DeleteResponse
		oldJobObj common.Job
	)
	//保存在etcd的任务key
	jobKey = common.ETCD_JOB_SAVE_DIR + name

	//从etcd中删除key
	if delResp, err = jobMgr.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}
	//从返回被删除的任务信息
	if len(delResp.PrevKvs) != 0{
		//把旧值解析一下 返回它
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		//赋值旧值
		oldJob = &oldJobObj
	}
	return
}

//获取任务列表
func (jobMgr *JobManager) ListJobs() (jobList []*common.Job, err error)  {
	var (
		jobDirKey   string
		getResp 	*clientv3.GetResponse
		kvPair 		*mvccpb.KeyValue
		job 		*common.Job
	)
	//保存在etcd任务的目录
	jobDirKey = common.ETCD_JOB_SAVE_DIR
	//获取目录下的所有任务
	if getResp, err = jobMgr.kv.Get(context.TODO(), jobDirKey, clientv3.WithPrefix()); err != nil {
		return
	}
	//初始化数组空间 len(jobList) == 0 而不是空
	jobList = make([]*common.Job, 0)

	//遍历所有任务,进行反序列化
	for _,kvPair = range getResp.Kvs {
		//getResp.Kvs值是个json 需要反序列化
		job = &common.Job{}
		if err = json.Unmarshal(kvPair.Value, job); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}
	return
}


//杀死任务
func (jobMgr *JobManager) KillJob(name string) (err error)  {
	//更新一下 key = /cron/killer/任务名
	var (
		jobKillerKey      string
		leaseId           clientv3.LeaseID
		leaseGrantResp    *clientv3.LeaseGrantResponse
	)

	//通知worker杀死对应任务
	jobKillerKey = common.ETCD_JOB_KILLER_DIR + name
	//让woker监听到一次put操作即可，创建一个租约让其稍后自动过期即可
	if leaseGrantResp, err = jobMgr.lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	//租约ID
	leaseId = leaseGrantResp.ID
	//设置killer标记
	if _,err = jobMgr.kv.Put(context.TODO(), jobKillerKey, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}
	return
}
