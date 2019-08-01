package master

import (
	"encoding/json"
	"fmt"
	"goDistributedCron/common"
	"net/http"
	"strconv"
)

//任务的HTTP接口 对外是一个struct
type ApiServer struct {
	httpServer *http.Server
}

var (
	//单例对象 如果想让其他包调用首字母必须大写
	G_apiServer *ApiServer
)

//保存任务接口
// POST job = {"name":"job1", "command":"echo hello", "cronExpr":"* * * * *"}
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
		bytes   []byte
	)
	//1. 解析post表单数据
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	//2. 取表单中的job字段
	postJob = req.PostForm.Get("job")
	//3. 反序列化job 保存job job传给-->jobmanager-->etcd
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}
	//4. 保存到etcd
	if oldJob, err = G_jobMgr.SaveJob(&job); err != nil {
		goto ERR
	}

	//5. 返回正常应答 ({"errno":0,"msg":"", "data":{....}})
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		//bytes是[]byte类型，转化成string类型便于查看
		//fmt.Println(string(bytes))
		resp.Write(bytes)
	}
	return
ERR:
	//6. 返回异常应答
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		//bytes是[]byte类型，转化成string类型便于查看
		//fmt.Println(string(bytes))
		resp.Write(bytes)
	}
}

//删除任务接口
//POST /job/delete  name=job1
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err    error //interface{}
		name   string
		oldJob *common.Job
		bytes  []byte
	)
	//POST : a=1&b=2&c=3
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	//获取删除的名称
	name = req.PostForm.Get("name")
	//去删除任务 调用jobManager删除功能
	if oldJob, err = G_jobMgr.DeleteJob(name); err != nil {
		goto ERR
	}

	//正常应答
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

//获取任务列表
func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var (
		jobList []*common.Job
		err     error
		bytes   []byte
	)

	//获取任务列表
	if jobList, err = G_jobMgr.ListJobs(); err != nil {
		fmt.Println(jobList)
		goto ERR
	}

	//正常应答
	if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
		//fmt.Println(string(bytes))
		resp.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		//fmt.Println(string(bytes))
		resp.Write(bytes)
	}
}

//强制杀死某个任务
// POST /job/kill name=job1
func handleJobKill(resp http.ResponseWriter, req *http.Request) {
	var (
		name  string
		err   error
		bytes []byte
	)

	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	//要杀死的任务名称
	name = req.PostForm.Get("name")

	//杀死任务
	if err = G_jobMgr.KillJob(name); err != nil {
		goto ERR
	}

	//正常应答
	if bytes, err = common.BuildResponse(0, "success", nil); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

// 查询任务日志
func handleJobLog(resp http.ResponseWriter, req *http.Request) {
	var (
		err        error
		name       string // 任务名字
		skipParam  string // 从第几条开始
		limitParam string // 返回多少条
		skip       int
		limit      int
		logArr     []*common.JobExecuteLog
		bytes      []byte
	)

	// 解析GET参数
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	// 获取请求参数 /job/log?name=job10&skip=0&limit=10
	name = req.Form.Get("name")
	skipParam = req.Form.Get("skip")
	limitParam = req.Form.Get("limit")
	if skip, err = strconv.Atoi(skipParam); err != nil {
		skip = 0
	}
	if limit, err = strconv.Atoi(limitParam); err != nil {
		limit = 20
	}

	if logArr, err = G_logManager.ListJobLog(name, skip, limit); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", logArr); err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

// 获取健康worker节点列表
func handleWorkerList(resp http.ResponseWriter, req *http.Request) {
	var (
		workerArr []string
		err       error
		bytes     []byte
	)

	//获取worker节点列表
	if workerArr, err = G_workerManager.ListWorkers(); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", workerArr); err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

//初始化ApiServer服务  HTTP服务
func InitApiServer() (err error) {
	//定义初始化变量
	var (
		staticDir     http.Dir     //静态文件根目录
		staticHandler http.Handler //静态文件的HTTP回调
	)

	//保存任务
	http.HandleFunc("/job/save", handleJobSave)
	//删除任务
	http.HandleFunc("/job/delete", handleJobDelete)
	//获取任务列表
	http.HandleFunc("/job/list", handleJobList)
	//杀死任务
	http.HandleFunc("/job/kill", handleJobKill)
	//查询日志
	http.HandleFunc("/job/log", handleJobLog)
	//worker节点监听
	http.HandleFunc("/worker/list", handleWorkerList)

	//静态文件目录
	staticDir = http.Dir(G_config.StaticWebRoot)
	//静态文件handler
	staticHandler = http.FileServer(staticDir)
	// /index.html   ./static_webroot/static/_webroot/index.html
	http.Handle("/", http.StripPrefix("/", staticHandler))

	//启动服务端
	go http.ListenAndServe(":"+strconv.Itoa(G_config.ApiPort), nil)

	return
}
