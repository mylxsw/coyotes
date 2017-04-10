# TaskRunner

![](./demo.gif)

## 命令参数

- **channel-default** string

    默认channel名称，用于消息队列 (default "default")

- **colorful-tty**

    是否启用彩色模式的控制台输出

- **concurrent** int

    并发执行线程数 (default 5)

- **host** string

    redis连接地址，必须指定端口(depressed,使用redis-host) (default "127.0.0.1:6379")

- **http-addr** string

    HTTP监控服务监听地址+端口 (default "127.0.0.1:60001")

- **password** string

    redis连接密码(depressed,使用redis-password)

- **pidfile** string

    pid文件路径 (default "/tmp/task-runner.pid")

- **redis-db** int

    redis默认数据库0-15

- **redis-host** string

    redis连接地址，必须指定端口 (default "127.0.0.1:6379")

- **redis-password** string

    redis连接密码

- **task-mode**

    是否启用任务模式，默认启用，关闭则不会执行消费 (default true)

## TODO

* [] 解决退出信号会传递到子进程的问题
* [] 队列channel持久化
* [*] 支持单个channel的退出
* [] 排队中的任务取消
* [] 实现队列调度器
* [] 使用可选后端记录任务执行结果、任务执行结果回调
* [] 实现RabbitMQ作为broker的支持
* [] 队列状态监控
* [] 集群支持（任务分发，新增队列）
* [] 支持脚本文件实时分发、执行
* [] 定时任务执行能力，发送一个任务到队列，设定一个执行时间，时间到了的时候再执行该任务（延迟执行）
