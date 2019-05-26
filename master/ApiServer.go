package master

import (
	"encoding/json"
	"go-cron-study/common"
	"net"
	"net/http"
	"strconv"
	"time"
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
		err      error
		postJob  string
		job      common.Job
		oldJob   *common.Job
		bytes    []byte
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
   if oldJob,err = G_jobMgr.SaveJob(&job); err != nil {
   	    goto ERR
   }
   //5. 返回正常应答 ({"errno":0,"msg":"", "data":{....}})
   if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
   	   resp.Write(bytes)
   }
   return
ERR:
	//6. 返回异常应答
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

//删除任务接口
//POST /job/delete  name=job1
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err error //interface{}
		name string
		oldJob *common.Job
		bytes []byte
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
     	goto ERR
	 }

	//正常应答
	if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}


//强制杀死某个任务
// POST /job/kill name=job1
func handleJobKill(resp http.ResponseWriter, req *http.Request) {
	var (
		name    string
		err     error
		bytes   []byte
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

//初始化ApiServer服务  HTTP服务
func InitApiServer() (err error) {
	//定义初始化变量
	var (
		mux           *http.ServeMux
		listener      net.Listener
		httpServer    *http.Server
	)
	//配置路由
	mux = http.NewServeMux()
	//保存任务
	mux.HandleFunc("/job/save", handleJobSave)
	//删除任务
	mux.HandleFunc("/job/delete", handleJobDelete)
	//获取任务列表
	mux.HandleFunc("/job/list", handleJobList)
	//杀死任务
	mux.HandleFunc("/job/kill", handleJobKill)


	//启动TCP监听 此处err不定义局部变量 直接返回值返回
	if listener, err = net.Listen("tcp", ":" + strconv.Itoa(G_config.ApiPort)); err != nil {
		return
	}

	//创建一个http服务
	httpServer = &http.Server{
		ReadTimeout: time.Duration(G_config.ApiReadTimeout),
		WriteTimeout:time.Duration(G_config.ApiWiterTimeout),
		Handler:mux,
	}

	//赋值单例
	G_apiServer = &ApiServer{
		httpServer:httpServer,
	}

	//启动服务端
	go httpServer.Serve(listener)

	return
}
