package worker

import (
	"goDistributedCron/common"
	"math/rand"
	"os/exec"
	"time"
)

//任务执行器
type JobExecutor struct {
}

//单例
var (
	G_executor *JobExecutor
)

//执行一个任务
func (excutor *JobExecutor) ExecuteJob(planInfo *common.JobExecutingPlan) {
	//启动一个协程来执行任务
	go func() {
		var (
			cmd                *exec.Cmd
			err                error
			outPutInfo         []byte
			result             *common.JobExecuteResult
			jobDistributedLock *JobDistributedLock
		)

		//拼装任务执行结果
		result = &common.JobExecuteResult{
			JobExecutePlanInfo: planInfo,
			OutPut:             make([]byte, 0),
		}

		//先获取分布式锁 然后在执行任务
		jobDistributedLock = G_jobMgr.CreateJobDistributedLock(planInfo.Job.Name)

		//任务执行开始时间
		result.StartTime = time.Now()

		//上分布式锁
		// 随机睡眠(0~1s) 保证多节点任务执行时间均衡 ntp保证微秒级时差
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		err = jobDistributedLock.TryJobDistributedLock()
		//锁完成后释放锁
		defer jobDistributedLock.UnLock()

		//当err不为空 表示上锁失败 然后赋值错误信息
		if err != nil {
			result.Err = err
			result.EndTime = time.Now()
		} else {
			//上锁是个网络事件，成功后 重新赋值任务初始时间
			result.StartTime = time.Now()

			//执行shell命令
			cmd = exec.CommandContext(planInfo.CancleCtx, "/bin/bash", "-c", planInfo.Job.Command)

			//执行并捕获输出结果
			outPutInfo, err = cmd.CombinedOutput()

			//任务执行结束时间
			result.EndTime = time.Now()
			result.OutPut = outPutInfo
			result.Err = err
		}

		//任务执行完成后，把执行的结果返回给Scheduler, Scheduler会从jobExcutingTableMap
		//中删除掉执行的记录信息
		G_scheduler.PushJobResult(result)
	}()
}

//启动执行器
func InitJobExecutor() (err error) {
	G_executor = &JobExecutor{}
	return
}
