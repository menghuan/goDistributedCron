package common

import "encoding/json"

//定时任务
type Job struct {
	//job任务名
	Name      string `json:"name"`
	//shell命令
	Command   string `json:"command"`
	//cron表达式
	CronExpr  string `json:"cronExpr"`
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
