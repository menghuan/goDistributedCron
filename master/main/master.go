package main

import (
	"flag"
	"fmt"
	"goDistributedCron/master"
	"runtime"
	"time"
)

var (
	//配置文件路径
	configFilePath string
)

//解析命令行参数
func initArgs() {
	//通过flag库来解析命令行参数 master -config ./master.json  xxx 123 -yyy ddd
	//master -h 查看帮助
	flag.StringVar(&configFilePath, "config", "./master.json", "指定master.json")
	//解析参数
	flag.Parse()
}

//初始化环境
func initEnv() {
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
	if err = master.InitConfig(configFilePath); err != nil {
		goto ERR
	}

	// 初始化服务发现管理器
	if err = master.InitWorkerManager(); err != nil {
		goto ERR
	}

	//初始化日志管理器
	if err = master.InitLogManager(); err != nil {
		goto ERR
	}

	//初始化任务管理器
	if err = master.InitJobManager(); err != nil {
		goto ERR
	}

	//启动ApiServer HTTP服务 增删改查
	if err = master.InitApiServer(); err != nil {
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
