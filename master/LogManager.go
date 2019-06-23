package master

import (
	"context"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"github.com/mongodb/mongo-go-driver/mongo/findopt"
	"goDistributedCron/common"
	"time"
)

// mongodb日志管理
type LogManager struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logManager *LogManager
)

//初始化日志管理器
func InitLogManager() (err error) {
	var (
		client *mongo.Client
	)

	// 建立mongodb连接
	if client, err = mongo.Connect(
		context.TODO(),
		G_config.MongodbUri,
		clientopt.ConnectTimeout(time.Duration(G_config.MongodbConnectTimeout)*time.Millisecond)); err != nil {
		return
	}

	G_logManager = &LogManager{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}
	return
}

// 从mongo中查看任务日志
func (logManager *LogManager) ListJobLog(name string, skip int, limit int) (logArr []*common.JobExecuteLog, err error) {
	var (
		filter  *common.JobExecuteLogFilter
		logSort *common.SortJobExecuteLogByStartTime
		cursor  mongo.Cursor
		jobLog  *common.JobExecuteLog
	)

	// len(logArr)
	logArr = make([]*common.JobExecuteLog, 0)

	// 过滤条件
	filter = &common.JobExecuteLogFilter{JobName: name}

	// 按照任务开始时间倒排
	logSort = &common.SortJobExecuteLogByStartTime{SortOrder: -1}

	// 查询
	if cursor, err = logManager.logCollection.Find(context.TODO(), filter, findopt.Sort(logSort), findopt.Skip(int64(skip)), findopt.Limit(int64(limit))); err != nil {
		return
	}
	// 延迟释放游标
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobExecuteLog{}

		// 反序列化BSON
		if err = cursor.Decode(jobLog); err != nil {
			continue // 有日志不合法
		}

		logArr = append(logArr, jobLog)
	}
	return
}
