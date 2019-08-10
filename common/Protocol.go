package common

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

//定时任务
type Job struct {
	//job任务名
	Name      string `json:"name"`
	//shell命令
	Command   string `json:"command"`
	//cron表达式
	CronExpr  string `json:"cronExpr"`
	//任务备注
	Remark  string `json:"remark"`
}

//etcd任务调度计划
type JobSchedulerPlan struct {
	//要调度的任务信息
	Job 		*Job
	//解析好的cron表达式
	Expr 		*cronexpr.Expression
	//下次调度执行时间
	NextTime  	time.Time
}

//etcd shell任务执行结果信息
type JobExecuteResult struct {
	//执行任务信息状态
	JobExecutePlanInfo  *JobExecutingPlan
	//执行脚本输出结果
	OutPut 			    []byte
	//脚本错误信息
	Err 				error
	//脚本启动时间
	StartTime			time.Time
	//脚本结果时间
	EndTime				time.Time
}


//etcd任务执行计划状态
type JobExecutingPlan struct {
	//要执行的任务信息
	Job 				*Job
	//理论调度执行时间
	PlanExecTime 		time.Time
	//实际调度执行时间
	RealExecTime  		time.Time
	//任务command的context
	CancleCtx			context.Context
	//用于取消command执行的cancle函数
	CancleFunc			context.CancelFunc
}

//任务执行日志
type JobExecuteLog struct {
	//任务名称
	JobName			string   	`bson:"jobName"`
	//脚本命令
	Command			string   	`bson:"command"`
	//错误原因
	Err				string   	`bson:"err"`
	//脚本输出
	Output			string   	`bson:"output"`
	//计划执行时间 毫秒
	PlanTime		int64   	`bson:"plantime"`
	//实际调度时间 毫秒  如果跟计划执行时间间隔很长 说明调度任务比较繁忙 一般情况在亚秒级
	ScheduleTime	int64		`bson:"scheduletime"`
	//任务执行开始时间 毫秒
	StartTime		int64   	`bson:"starttime"`
	//任务执行结束时间 毫秒
	EndTime			int64   	`bson:"endtime"`
}

// 日志批次
type JobExecuteLogBatch struct {
	// 多条日志
	Logs []interface{}
}

// 任务日志过滤条件
type JobExecuteLogFilter struct {
	JobName string `bson:"jobName"`
}

// 任务日志排序规则
type SortJobExecuteLogByStartTime struct {
	SortOrder int `bson:"startTime"`	// {startTime: -1}
}


//etcd变化事件
type JobEvent struct {
	//SAVE, DELETE
	EventType int
	//任务信息
	Job       *Job
}


//HTTP接口应答
type Response struct {
	//错误码
	Errno     int `json:"errno"`
	//提示信息
	Msg       string `json:"msg"`
	//返回值
	Data      interface{} `json:"data"`
}

//组织应答方法
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	//1. 定义一个response
	var (
		 response Response
	)

	//2. 赋值信息到response对象中
	response.Errno  = errno
	response.Msg    = msg
	response.Data   = data

	//3. 序列化json
	resp, err = json.Marshal(response)
	return
}

//反序列化job
func UnpackJob(value []byte) (ret *Job, err error)  {
	var (
		job *Job
	)
	job = &Job{}
	if 	err = json.Unmarshal(value, job); err != nil {
		return
	}

	ret = job
	return
}

//从etcd的key中提取任务名  /distributed_cron/jobs/jobxx  删掉/distributed_cron/jobs 得到jobxx
func ExtractJobName(jobKey string) string  {
	return strings.TrimPrefix(jobKey, ETCD_JOB_SAVE_DIR)
}

//从etcd的key中提取强杀的任务名  /distributed_cron/killers/jobxx  删掉/distributed_cron/jobs 得到jobxx
func ExtractKillerJobName(jobKillerKey string) string  {
	return strings.TrimPrefix(jobKillerKey, ETCD_JOB_KILLER_DIR)
}

//任务变化事件封装  主要有2种 1）更新任务 2）删除任务
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent)  {
	return &JobEvent{
		EventType:eventType,
		Job:job,
	}
}

//构建调度任务计划
func BuildJobSchedulerPlan(job *Job) (jobSchedulerPlan 	*JobSchedulerPlan, err error)   {
	var (
		expr *cronexpr.Expression
	)

	//解析job的Cron表达式
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	//生成任务调度计划对象
	jobSchedulerPlan = &JobSchedulerPlan{
		Job:job,
		Expr:expr,
		NextTime:expr.Next(time.Now()),  //通过cron表达式的下一次执行获取下一次执行时间
	}
	return
}


//构建执行状态信息
func BuildJobExcutingPlan(jobSchedulerPlan *JobSchedulerPlan) (jobExecutingPlan *JobExecutingPlan) {
	jobExecutingPlan = &JobExecutingPlan{
		Job:jobSchedulerPlan.Job,
		PlanExecTime:jobSchedulerPlan.NextTime, //计划调度执行时间
		RealExecTime:time.Now(), //实际调度执行时间
	}
	//获取执行任务的上下文和取消函数
	jobExecutingPlan.CancleCtx, jobExecutingPlan.CancleFunc = context.WithCancel(context.TODO())
	return
}

// 提取worker的IP
func ExtractWorkerIP(regKey string) (string) {
	return strings.TrimPrefix(regKey, ETCD_JOB_WORKER_DIR)
}
