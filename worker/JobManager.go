package worker

import (
	"context"
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
	watcher  clientv3.Watcher
}

var (
	//任务管理器单例对象
	G_jobMgr *JobManager
)

//初始化任务管理器 主要是监听etcd任务变化
func InitJobManager() (err error){
	var (
		config    clientv3.Config
		client    *clientv3.Client
		kv        clientv3.KV
		lease     clientv3.Lease
		watcher   clientv3.Watcher
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
	watcher = clientv3.NewWatcher(client)

	//赋值单例
	G_jobMgr = &JobManager{
		client:client,
		kv:kv,
		lease:lease,
		watcher:watcher,
	}

	//启动任务监听
	G_jobMgr.watchJob()

	//启动任务监听强杀
	G_jobMgr.watchJobKiller()

	return
}

//监听etcd任务变化
func(jobMgr *JobManager) watchJob() (err error){
	var (
		getResp 			*clientv3.GetResponse
		kvPair      		*mvccpb.KeyValue
		job					*common.Job
		watchStartRevision  int64
		watchChan           clientv3.WatchChan //底层是channel
		watchResp           clientv3.WatchResponse
		watchEvent          *clientv3.Event
		jobName				string
		jobEvent            *common.JobEvent
	)

	//1.获取一下etcd中/distributed_cron/jobs目录下的所有任务，并且获知当前集群的revision
	if getResp,err = jobMgr.kv.Get(context.TODO(), common.ETCD_JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	//循环当前所有的任务
	for _,kvPair = range getResp.Kvs {
		//反序列化json得到job
		if job, err = common.UnpackJob(kvPair.Value); err == nil {
			jobEvent = common.BuildJobEvent(common.ETCD_JOB_EVENT_SAVE, job)
			//TODO:	要把这个任务job同步给调度协程（scheduler）
			G_scheduler.pushJobEvent(jobEvent)
		}
	}
	//2.从该revision向后监听变化的事件
	go func() { //启动监听协程
	   //从Get时刻的后续版本开始监听
	   watchStartRevision = getResp.Header.Revision + 1
	   //监听/distributed_cron/jobs/目录的后续变化 需要版本和withPrefix
	   watchChan = jobMgr.watcher.Watch(context.TODO(), common.ETCD_JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
	   //处理监听事件
	   for watchResp = range watchChan {
		   for _, watchEvent = range watchResp.Events{
			   switch watchEvent.Type {
			   case mvccpb.PUT:  //任务保存更新事件
				   //反序列化job，一旦反序列化解析失败后跳过
				   if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
					  continue
				   }
				   //构建一个更新event事件
				   jobEvent = common.BuildJobEvent(common.ETCD_JOB_EVENT_SAVE, job)
			   case mvccpb.DELETE: //任务删除事件
				   //Delete /distributed_cron/jobs/jobxx
				   jobName = common.ExtractJobName(string(watchEvent.Kv.Value))
				   //通过job集合中获取job任务
				   job = &common.Job{Name:jobName}
				   //构建一个删除event事件
				   jobEvent = common.BuildJobEvent(common.ETCD_JOB_EVENT_DELETE, job)
			   }
			   //TODO:推送事件给scheduler调度协程
			   G_scheduler.pushJobEvent(jobEvent)
		   }
	   }
	}()

	return
}


//监听etcd强杀任务通知
func(jobMgr *JobManager) watchJobKiller() (err error){
	var (
		watchChan           clientv3.WatchChan //底层是channel
		watchResp           clientv3.WatchResponse
		watchEvent          *clientv3.Event
		jobName				string
		jobEvent            *common.JobEvent
		job					*common.Job
	)

	//监听/distributed_cron/killers目录
	go func() { //启动监听协程
		//监听/distributed_cron/killers/目录的后续变化 需要版本和withPrefix
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.ETCD_JOB_KILLER_DIR, clientv3.WithPrefix())
		//处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events{
				switch watchEvent.Type {
				case mvccpb.PUT:  //杀死强杀任务
					jobName = common.ExtractKillerJobName(string(watchEvent.Kv.Key))
					//任务名称
					job = &common.Job{Name:jobName}
					jobEvent = common.BuildJobEvent(common.ETCD_JOB_EVENT_KILL, job)
					//TODO:推送事件给scheduler调度协程
					G_scheduler.pushJobEvent(jobEvent)
				case mvccpb.DELETE: //killers标记过期，被自动删除
				}
			}
		}
	}()

	return
}


//创建任务执行分布式锁
func (jobMgr *JobManager) CreateJobDistributedLock(jobName string) (jobDistributedLock *JobDistributedLock)  {
	//返回锁
	jobDistributedLock = InitJobDistributedLock(jobName, jobMgr.kv, jobMgr.lease)
	return
}
