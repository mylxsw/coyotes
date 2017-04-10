# TaskRunner

TaskRunner的诞生起源于在使用Laravel的定时任务时，由于PHP本身的限制（不安装线程相关扩展），无法实现并发的任务执行，如果任务执行时间过长，就会影响到其它定时任务的执行。不同于其它重量级任务队列，TaskRunner仅仅提供了对命令行程序执行的支持，这样就避免了开发者需要学习任务队列相关API，针对任务队列开发任务程序的需要。只需要提供一个可执行的文件或者脚本执行命令，TaskRunner就可以并发的执行。

简单的说，就是“你把命令给我，我来执行，仅此而已”。

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

## 安装部署

编译安装需要安装 **Go1.7+**，执行以下命令编译

    make build-mac

上述命令编译后是当前平台的可执行文件（**./bin/task-runner**）。比如在Mac系统下完成编译后只能在Mac系统下使用，Linux系统下编译则可以在Linux系统下使用。
如果要在Mac系统下编译Linux系统下使用的可执行文件，需要本地先配置好交叉编译选线，之后执行下面的命令完成Linux版本编译

    make build-linux

将生成的执行文件（在**bin**目录）复制到系统的`/usr/local/bin`目录即可。

    mv ./bin/task-runner /usr/local/bin/task-runner

项目目录下提供了`supervisor.conf`配置文件可以直接在supervisor下使用，使用supervisor管理TaskRunner进程。TaskRunner在启动之前需要确保Redis服务是可用的。

    /usr/local/bin/task-runner -redis-host 127.0.0.1:6379 -password REDIS访问密码 

如果需要退出进程，需要向进程发送`USR2`信号，以实现平滑退出。

    kill -USR2 $(pgrep task-runner)

> 请不要直接使用`kill -9`终止进程，使用它会强制关闭进程，可能会导致正在执行中的命令被终止，造成任务队列数据不完整。

## 任务推送方式

将待执行的任务推送给TaskRunner执行有两种方式

- 直接将任务写入到Redis的队列`task:prepare:queue`
- 使用HTTP API

### 直接写入Redis队列

将任务以json编码的形式写入到Redis的`task:prepare:queue`即可。

    $redis->lpush('task:prepare:queue', json_encode([
      'task' => $taskName,
      'chan' => $channel,
      'ts'   => time(),
    ], JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE))

写入到`task:prepare:queue`之后，TaskRunner会实时的去从中取出任务分发到相应的channel队列供worker消费。

### 使用HTTP API

Request:

    POST /channels/default HTTP/1.1
    Accept: */*
    Host: localhost:60001
    content-type: multipart/form-data; boundary=--------------------------019175029883341751119913
    content-length: 179
    
    ----------------------------019175029883341751119913
    Content-Disposition: form-data; name="task"
    
    ping -c 40 baidu.com
    ----------------------------019175029883341751119913--

    
Response:
    
    HTTP/1.1 200 OK
    Content-Type: application/json; charset=utf-8
    Date: Mon, 10 Apr 2017 13:05:56 GMT
    Content-Length: 92
    
    {"status_code":200,"message":"ok","data":{"task_name":"ping -c 40 baidu.com","result":true}}


## HTTP API

TaskRunner提供了Restful风格的API用于对其进行管理。

[![Run in Postman](https://run.pstmn.io/button.svg)](https://app.getpostman.com/run-collection/21728c33bdd4b4b703d0)

### **GET /channels** 查询所有channel的状态

#### 请求参数

无

#### 响应结果示例

    {
      "status_code": 200,
      "message": "ok",
      "data": {
        "biz": {
          "tasks": [],
          "count": 0
        },
        "cron": {
          "tasks": [],
          "count": 0
        },
        "default": {
          "tasks": [
            {
              "task_name": "ping -c 40 baidu.com",
              "channel": "default",
              "status": "running"
            }
          ],
          "count": 1
        },
        "test": {
          "tasks": [],
          "count": 0
        }
      }
    }

任务状态字段`status`对应关系


| status | 说明 |
| --- | --- |
| expired | 已过期（应该不会出现，待确认）  |
| queued | 已队列，任务正在等待执行 |
| running | 任务执行中 |


### **GET /channels/{channel_name}** 查新某个channel的状态

#### 请求参数

无

#### 响应结果示例

    {
      "status_code": 200,
      "message": "ok",
      "data": {
        "tasks": [],
        "count": 0
      }
    }

### **POST /channels/{channel_name}** 推送任务到任务队列

#### 请求参数

| 参数 | 说明 |
| --- | --- |
| task | 任务命令，比如`date` |

#### 响应结果示例

    {
      "status_code": 200,
      "message": "ok",
      "data": {
        "task_name": "date",
        "result": true
      }
    }

### **POST /channels** 新建任务队列

#### 请求参数

| 参数 | 说明 |
| --- | --- |
| name | 队列名称，也就是推送任务时的channel |
| distinct | 是否对命令进行去重，可选值为true/false，默认为false |
| worker | worker数量 |

#### 响应结果示例

    {
      "status_code": 200,
      "message": "ok",
      "data": null
    }

## TODO

* [x] 解决退出信号会传递到子进程的问题
* [ ] 队列channel持久化
* [x] 支持单个channel的退出
* [ ] 排队中的任务取消
* [ ] 实现队列调度器
* [ ] 使用可选后端记录任务执行结果、任务执行结果回调
* [ ] 实现RabbitMQ作为broker的支持
* [ ] 队列状态监控
* [ ] 集群支持（任务分发，新增队列）
* [ ] 支持脚本文件实时分发、执行
* [ ] 定时任务执行能力，发送一个任务到队列，设定一个执行时间，时间到了的时候再执行该任务（延迟执行）


