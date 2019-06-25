# goDistributedCron
一个简单的分布式任务调度系统，有简单的后台可以操作

# 后续会持续完善跟进
  ## master  任务的管理以及一个后台操作管理
    主程序初始化相关操作
    1. 线程数设置
    2. 命令行参数统一解析
    2. 配置文件处理
    3. 任务事件管理器 任务的列表，保存，删除，杀死等功能实现，任务保存在ETCD中
    4. http服务 路由配置 路由逻辑实现
    5. 返回值统一处理

  ## worker
    抢占任务 任务调度 执行任务
    1. 从etcd中把任务job同步到内存中
    2. 实现调度模块，基于cron表达式解析并调度job
    3. 实现执行模块，并发的执行多个job
    4. 实现分布式锁，解决并发惊群调度问题
    5. 把执行的日志保存到存储中
	
	
# 系统简单架构图如下:
   <p align="center"><img src="https://github.com/menghuan/goDistributedCron/doc/go_cron_arch.png.png" width="800px"></p>
