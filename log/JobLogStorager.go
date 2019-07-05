package log

import (
	"context"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"goDistributedCron/common"
	"goDistributedCron/worker"
	"time"
)

//暂时先用MongoDB存储日志 后续会换成es
type JobLogStorager struct {
	//mongo客户端
	client *mongo.Client
	//mongo集合
	logCollection *mongo.Collection
	//日志channel队列
	logChan chan *common.JobExecuteLog
	//自动提交chan
	autoCommitChan chan *common.JobExecuteLogBatch
}

//单例
var (
	G_jobLogStorager *JobLogStorager
)

// 批量写入日志
func (jobLogStorager *JobLogStorager) saveJobLogs(batch *common.JobExecuteLogBatch) {
	jobLogStorager.logCollection.InsertMany(context.TODO(), batch.Logs)
}

//日志存储协程
func (jobLogStorager *JobLogStorager) writeJobLogLoop() {
	var (
		log          *common.JobExecuteLog
		logBatch     *common.JobExecuteLogBatch // 当前的批次
		commitTimer  *time.Timer
		timeoutBatch *common.JobExecuteLogBatch // 超时批次
	)

	for {
		select {
		case log = <-jobLogStorager.logChan:
			//每次插入都需要等待一次mongodb网络请求往返，耗时可能因为网络慢花费比较长的时间
			if logBatch == nil {
				logBatch = &common.JobExecuteLogBatch{}
				// 让这个批次超时自动提交(给1秒的时间）
				commitTimer = time.AfterFunc(
					time.Duration(worker.G_config.JobLogCommitTimeout)*time.Millisecond,
					func(batch *common.JobExecuteLogBatch) func() {
						return func() {
							jobLogStorager.autoCommitChan <- batch
						}
					}(logBatch),
				)
			}

			// 把新日志追加到批次中
			logBatch.Logs = append(logBatch.Logs, log)

			// 如果批次满了, 就立即发送
			if len(logBatch.Logs) >= worker.G_config.JobLogBatchSize {
				// 发送日志
				jobLogStorager.saveJobLogs(logBatch)
				// 清空logBatch
				logBatch = nil
				// 取消定时器
				commitTimer.Stop()
			}
		case timeoutBatch = <-jobLogStorager.autoCommitChan: // 过期的批次
			// 判断过期批次是否仍旧是当前的批次
			if timeoutBatch != logBatch {
				continue // 跳过已经被提交的批次
			}
			// 把批次写入到mongo中
			jobLogStorager.saveJobLogs(timeoutBatch)
			// 清空logBatch
			logBatch = nil
		}
	}
}

//初始化日志存储
func InitJobLogStorager() (err error) {
	var (
		client *mongo.Client
	)

	// 建立mongodb连接
	if client, err = mongo.Connect(
		context.TODO(),
		worker.G_config.MongodbUri,
		clientopt.ConnectTimeout(time.Duration(worker.G_config.MongodbConnectTimeOut)*time.Microsecond)); err != nil {
		return
	}

	// 选择db和collection
	G_jobLogStorager = &JobLogStorager{
		client:         client,
		logCollection:  client.Database("cron").Collection("log"),
		logChan:        make(chan *common.JobExecuteLog, 1000),
		autoCommitChan: make(chan *common.JobExecuteLogBatch, 1000),
	}

	// 启动一个mongodb处理协程
	go G_jobLogStorager.writeJobLogLoop()

	return
}

// 发送日志
func (jobLogStorager *JobLogStorager) AppendJobLogs(jobExecuteLog *common.JobExecuteLog) {
	select {
	case jobLogStorager.logChan <- jobExecuteLog:
	default:
		// 队列满了就丢弃
	}
}
