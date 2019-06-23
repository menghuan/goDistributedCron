package common

const (
	//etcd任务保存目录
	ETCD_JOB_SAVE_DIR = "/distributed_cron/jobs/"

	//etcd任务杀死目录
	ETCD_JOB_KILLER_DIR = "/distributed_cron/killers/"

	//etcd任务锁目录
	ETCD_JOB_LOCK_DIR = "/distributed_cron/locks/"

	//保存更新任务事件
	ETCD_JOB_EVENT_SAVE = 1

	//删除任务事件
	ETCD_JOB_EVENT_DELETE  = 2

	//强杀任务事件
	ETCD_JOB_EVENT_KILL = 3
)
