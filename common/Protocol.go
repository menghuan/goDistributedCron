package common

import (
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

//从etcd的key中提取任务名  /cron/jobs/jobxx  删掉/cron/jobs 得到jobxx
func ExtractJobName(jobKey string) string  {
	return strings.TrimPrefix(jobKey, ETCD_JOB_SAVE_DIR)
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
	return
}
