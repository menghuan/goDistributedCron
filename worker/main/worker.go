package main

import (
	"flag"
	"fmt"
	"goDistributedCron/worker"
	"runtime"
	"time"
)
var (
	//配置文件路径
	configFilePath string
)
//解析命令行参数
func initArgs()  {
	//通过flag库来解析命令行参数 worker -config ./worker.json
	//worker -h 查看帮助
	flag.StringVar(&configFilePath, "config", "./worker.json", "指定worker.json")
	//解析参数
	flag.Parse()
}

//初始化环境
func initEnv()  {
	//初始化线程  线程数跟cpu核心一样 性能最好
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)
	//初始化命令行参数
	initArgs()

	//初始化线程
	initEnv()

	//加载配置文件 通过传递文件名 进行读取 反序列化解析
	if err = worker.InitConfig(configFilePath); err != nil {
		goto ERR
	}

	//启动服务注册发现管理器
	if err = worker.InitJobRegister(); err != nil {
		goto ERR
	}

	//启动日志协程
	if err = worker.InitJobLogStorager(); err != nil {
		goto ERR
	}

	//初始化任务执行器
	if err = worker.InitJobExecutor(); err != nil {
		goto ERR
	}

	//初始化任务调度器 会监听事件，有监听事件会从JobManager同步给Scheduler
	if err = worker.InitJobScheduler(); err != nil {
		goto ERR
	}

	//初始化任务管理器
	if err = worker.InitJobManager(); err != nil {
		goto ERR
	}

	//永远不停止 这样的话就可以常驻了
	for {
		time.Sleep(1 * time.Second)
	}

	return

ERR:
	fmt.Println(err)
}
