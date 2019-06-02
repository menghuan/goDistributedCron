package worker

import (
	"context"
	"goDistributedCron/common"
	"os/exec"
	"time"
)

//任务执行器
type JobExecutor struct {

}

//单例
var (
	G_executor  *JobExecutor
)

//执行一个任务
func (excutor *JobExecutor) ExecuteJob(planInfo *common.JobExecutingPlan) {
	//启动一个协程来执行任务
	go func() {
		var (
			cmd     		*exec.Cmd
			err     		error
			outPutInfo  	[]byte
			result			*common.JobExecuteResult
		)

		//拼装任务执行结果
		result = &common.JobExecuteResult{
			JobExecutePlanInfo:planInfo,
			OutPut:make([]byte, 0),
		}

		//任务执行开始时间
		result.StartTime = time.Now()

		//执行shell命令
		cmd = exec.CommandContext(context.TODO(), "/bin/bash", "-c", planInfo.Job.Command)

		//执行并捕获输出结果
		outPutInfo, err = cmd.CombinedOutput()

		//任务执行结束时间
		result.EndTime = time.Now()
		result.OutPut = outPutInfo
		result.Err = err

		//任务执行完成后，把执行的结果返回给Scheduler, Scheduler会从jobExcutingTableMap中删除掉执行的记录信息
		G_scheduler.PushJobResult(result)
	}()
}

//启动执行器
func InitJobExecutor() (err error)  {
	G_executor = &JobExecutor{

	}
	return
}
