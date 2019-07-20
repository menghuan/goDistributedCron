package worker

import (
	"fmt"
	"goDistributedCron/common"
	"time"
)

//任务调度
type JobScheduler struct {
	//etcd任务队列
	jobEventChan chan *common.JobEvent
	//etcd任务执行计划表
	jobPlanTableMap map[string]*common.JobSchedulerPlan
	//etcd任务执行调度表
	jobExcutingTableMap map[string]*common.JobExecutingPlan
	//任务结果队列
	jobExecuteResultChan chan *common.JobExecuteResult
}

//单例 
var (
	G_scheduler *JobScheduler
)

//处理任务事件
func (scheduler *JobScheduler) HandleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulerPlan   *common.JobSchedulerPlan
		jobExisted         bool
		jobExecutePlanInfo *common.JobExecutingPlan
		jobExecuting       bool
		err                error
	)
	switch jobEvent.EventType {
	//保存修改任务事件
	case common.ETCD_JOB_EVENT_SAVE:
		//如果cron表达式解析失败
		if jobSchedulerPlan, err = common.BuildJobSchedulerPlan(jobEvent.Job); err != nil {
			return
		}
		//将执行计划加入到调度任务计划表中
		scheduler.jobPlanTableMap[jobEvent.Job.Name] = jobSchedulerPlan
	//删除任务事件
	case common.ETCD_JOB_EVENT_DELETE:
		//判断调度任务计划表是否还存在该任务 存在则删除
		if jobSchedulerPlan, jobExisted = scheduler.jobPlanTableMap[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTableMap, jobEvent.Job.Name)
		}
	//强杀任务事件
	case common.ETCD_JOB_EVENT_KILL:
		//先判断任务是否在执行中，取消掉command执行
		if jobExecutePlanInfo, jobExecuting = scheduler.jobExcutingTableMap[jobEvent.Job.Name]; jobExecuting {
			//如果在执行 触发command杀死任务子进程，任务得到退出
			jobExecutePlanInfo.CancleFunc()
		}
	}

}

//处理任务结果
func (scheduler *JobScheduler) HandleJobResult(result *common.JobExecuteResult) {
	var (
		jobExecuteLog *common.JobExecuteLog
	)
	//删除执行状态
	delete(scheduler.jobExcutingTableMap, result.JobExecutePlanInfo.Job.Name)

	//生成执行日志
	if result.Err != common.ERR_DISTRIBUTED_ALREADY_LOCK {
		//当不是因为锁被占用而失败的情况下，锁被占用是正常情况 跳过
		jobExecuteLog = &common.JobExecuteLog{
			JobName:      result.JobExecutePlanInfo.Job.Name,
			Command:      result.JobExecutePlanInfo.Job.Command,
			Output:       string(result.OutPut),
			PlanTime:     result.JobExecutePlanInfo.PlanExecTime.UnixNano() / 1000 / 1000, //毫秒
			ScheduleTime: result.JobExecutePlanInfo.RealExecTime.UnixNano() / 1000 / 1000, //毫秒
			StartTime:    result.StartTime.UnixNano() / 1000 / 1000,                       //毫秒
			EndTime:      result.EndTime.UnixNano() / 1000 / 1000,                         //毫秒
		}

		if result.Err != nil {
			jobExecuteLog.Err = result.Err.Error() //错误信息
		} else {
			jobExecuteLog.Err = ""
		}
		//TODO: 存储日志到mongo中 需要在另一个协程中执行 不要影响schedulerCheckLoop调度任务检测
		G_jobLogStorager.AppendJobLogs(jobExecuteLog)
	}
	//打印结果信息
	fmt.Println("任务执行完成:", result.JobExecutePlanInfo.Job.Name, result.OutPut, result.Err)
}

//尝试执行任务 调度和执行
func (scheduler *JobScheduler) TryStartJob(jobPlan *common.JobSchedulerPlan) {
	var (
		jobExcutingPlan *common.JobExecutingPlan
		jobExcuting     bool
	)

	//执行的任务可能会运行很久，1分钟调度60次，但是只能执行1次，防止并发问题出现
	//如果任务正在执行 跳过本地调度
	if jobExcutingPlan, jobExcuting = scheduler.jobExcutingTableMap[jobPlan.Job.Name]; jobExcuting {
		fmt.Println("任务尚未退出，跳过执行:", jobPlan.Job.Name)
		return
	}

	//构建执行状态计划信息
	jobExcutingPlan = common.BuildJobExcutingPlan(jobPlan)

	//保存执行状态
	scheduler.jobExcutingTableMap[jobPlan.Job.Name] = jobExcutingPlan

	//打印任务信息
	fmt.Println("正在执行任务:", jobExcutingPlan.Job.Name, jobExcutingPlan.PlanExecTime, jobExcutingPlan.RealExecTime)

	//开始执行任务
	G_executor.ExecuteJob(jobExcutingPlan)
}

//重新计算任务调度状态 jobSchedulerPlan *common.JobSchedulerPlan
func (scheduler *JobScheduler) RecalculateScheduler() (schedulerAfterTime time.Duration) {
	var (
		jobPlan  *common.JobSchedulerPlan //任务调度计划
		now      time.Time                //当前时间
		nearTime *time.Time               //一个更近的时间
	)
	//如果任务计划表为空的话 随便休眠1秒
	if len(scheduler.jobPlanTableMap) == 0 {
		schedulerAfterTime = 1 * time.Second
		return
	}

	//获取当前时间
	now = time.Now()
	//1. 循环遍历所有任务
	for _, jobPlan = range scheduler.jobPlanTableMap {
		//当执行计划任务的下次执行时间小于或者等于当前时间的话 就尝试执行任务
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			fmt.Println("调度执行任务:", jobPlan.Job.Name)
			//TODO: 尝试执行任务  可能上一次还没结束 不一定能启动成功
			scheduler.TryStartJob(jobPlan)
			//更新计划的下一次时间
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}
		//统计最近一个要过期的任务时间（N秒后即将过期 == schedulerAfterTime 就休息N秒）
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	// 下次调度时间 （最近要执行任务的时间 - 当前时间） 相当于休眠N秒
	schedulerAfterTime = (*nearTime).Sub(now)
	return
}

//调度协程检测
func (scheduler *JobScheduler) SchedulerCheckLoop() {
	var (
		jobEvent           *common.JobEvent
		schedulerAfterTime time.Duration
		schedulerTimer     *time.Timer              //监听器
		jobExcuteResult    *common.JobExecuteResult //执行结果
	)
	//初始化任务调度状态监测（1秒）
	schedulerAfterTime = scheduler.RecalculateScheduler()

	//调度的延时定时器
	schedulerTimer = time.NewTimer(schedulerAfterTime)

	//定时任务 common.Job
	for {
		select {
		//循环监听任务变化事件 包括添加 修改 强杀任务
		case jobEvent = <-scheduler.jobEventChan:
			//对内存chan中的任务做增删改查操作
			G_scheduler.HandleJobEvent(jobEvent)
		//任务到期了
		case <-schedulerTimer.C:
		//监听任务执行结果
		case jobExcuteResult = <-scheduler.jobExecuteResultChan:
			G_scheduler.HandleJobResult(jobExcuteResult)
		}
		//调度一次任务
		schedulerAfterTime = scheduler.RecalculateScheduler()
		//重置任务间隔
		schedulerTimer.Reset(schedulerAfterTime)
	}
}

//push任务事件到调度协程器
func (scheduler *JobScheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

//执行完结果后 回传任务执行结果给Scheduler
func (scheduler *JobScheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	scheduler.jobExecuteResultChan <- jobResult
}

//初始化任务调度器
func InitJobScheduler() (err error) {
	G_scheduler = &JobScheduler{
		//定义一个1000容量的chan队列
		jobEventChan: make(chan *common.JobEvent, 1000),
		//定义一个任务调度计划表
		jobPlanTableMap: make(map[string]*common.JobSchedulerPlan),
		//定义一个任务执行计划表
		jobExcutingTableMap: make(map[string]*common.JobExecutingPlan),
		//定义一个任务执行结果
		jobExecuteResultChan: make(chan *common.JobExecuteResult, 1000),
	}

	//启动调度协程
	go G_scheduler.SchedulerCheckLoop()
	return
}

