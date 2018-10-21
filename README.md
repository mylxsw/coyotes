# Coyotes

不同于其它重量级任务队列，Coyotes仅仅提供了对命令行程序执行的支持，这样就避免了开发者需要学习任务队列相关API，针对任务队列开发任务程序的需要。只需要提供一个可执行的文件或者脚本执行命令，Coyotes就可以并发的执行。

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

    pid文件路径 (default "/tmp/coyotes.pid")

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

上述命令编译后是当前平台的可执行文件（**./bin/coyotes**）。比如在Mac系统下完成编译后只能在Mac系统下使用，Linux系统下编译则可以在Linux系统下使用。
如果要在Mac系统下编译Linux系统下使用的可执行文件，需要本地先配置好交叉编译选线，之后执行下面的命令完成Linux版本编译

    make build-linux

将生成的执行文件（在**bin**目录）复制到系统的`/usr/local/bin`目录即可。

    mv ./bin/coyotes /usr/local/bin/coyotes

项目目录下提供了`supervisor.conf`配置文件可以直接在supervisor下使用，使用supervisor管理Coyotes进程。Coyotes在启动之前需要确保Redis服务是可用的。

    /usr/local/bin/coyotes -redis-host 127.0.0.1:6379 -password REDIS访问密码 

如果需要退出进程，需要向进程发送`USR2`信号，以实现平滑退出。

    kill -USR2 $(pgrep coyotes)

> 请不要直接使用`kill -9`终止进程，使用它会强制关闭进程，可能会导致正在执行中的命令被终止，造成任务队列数据不完整。

## 任务推送方式

将待执行的任务推送给Coyotes执行有两种方式

- 直接将任务写入到Redis的队列`task:prepare:queue`
- 使用HTTP API

### 直接写入Redis队列

将任务以json编码的形式写入到Redis的`task:prepare:queue`即可。

    $redis->lpush('task:prepare:queue', json_encode([
      'task' => $taskName,
      'chan' => $channel,
      'ts'   => time(),
    ], JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE))

写入到`task:prepare:queue`之后，Coyotes会实时的去从中取出任务分发到相应的channel队列供worker消费。

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

Coyotes提供了Restful风格的API用于对其进行管理。

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

### **POST /channels/{channel_name}/tasks** 推送任务到任务队列

#### 请求参数

| 参数 | 说明 |
| --- | --- |
| task | 任务名称，不指定command参数时，作为要执行的命令，指定command参数时则仅仅作为任务名称展示使用 |
| command | 要执行的命令，推荐使用该方式，如果不指定，则使用task参数作为要执行的命令 |
| args | 命令参数，仅在指定command参数时有效，可以指定多个，依次作为命令参数 |
| delay | 延迟执行，单位为秒，默认为0 |

#### 请求示例

    POST /channels/default/tasks HTTP/1.1
    cache-control: no-cache
    Postman-Token: c084444b-6657-4d30-8128-7ad5d898b69b
    Content-Type: multipart/form-data; boundary=--------------------------488585830099792470732554
    User-Agent: PostmanRuntime/3.0.11-hotfix.2
    Accept: */*
    Host: localhost:60001
    accept-encoding: gzip, deflate
    content-length: 700
    Connection: keep-alive

    ----------------------------488585830099792470732554
    Content-Disposition: form-data; name="task"

    pwd
    ----------------------------488585830099792470732554
    Content-Disposition: form-data; name="command"

    ping
    ----------------------------488585830099792470732554
    Content-Disposition: form-data; name="args"

    -c
    ----------------------------488585830099792470732554
    Content-Disposition: form-data; name="args"

    10
    ----------------------------488585830099792470732554
    Content-Disposition: form-data; name="args"

    baidu.com
    ----------------------------488585830099792470732554
    Content-Disposition: form-data; name="delay"

    10
    ----------------------------488585830099792470732554--
    HTTP/1.1 200 OK
    Content-Type: application/json; charset=utf-8
    Date: Tue, 16 May 2017 13:11:31 GMT
    Content-Length: 124

    {"status_code":200,"message":"ok","data":{"task_id":"2fe1347e-d0f1-4da7-a08e-0f299e812008","task_name":"pwd","result":true}}


#### 响应结果示例

    {
      "status_code": 200,
      "message": "ok",
      "data": {
        "task_id": "2fe1347e-d0f1-4da7-a08e-0f299e812008",
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

### **GET /channels/{channel_name}/failed-tasks** 查询channel下所有失败的任务

#### 响应结果示例

    {
        "status_code": 200,
        "message": "ok",
        "data": {
            "3688442d-a896-4456-984e-70aae8a75139": {
                "task_id": "3688442d-a896-4456-984e-70aae8a75139",
                "task_name": "git-command",
                "command": {
                    "name": "git",
                    "args": null
                },
                "channel": "default",
                "status": "",
                "retry_count": 3,
                "failed_at": "2017-10-17T16:49:09.647182+08:00"
            }
        }
    }

### **GET /channels/{channel_name}/failed-tasks/{task_id}** 查询channel下失败的某个任务

#### 请求参数

| 参数 | 说明 |
|---|---|
| channel_name | 任务所在的channel |
| task_id | 任务ID |

#### 响应示例

    {
        "status_code": 200,
        "message": "ok",
        "data": {
            "task_id": "3688442d-a896-4456-984e-70aae8a75139",
            "task_name": "git-command",
            "command": {
                "name": "git",
                "args": null
            },
            "channel": "default",
            "status": "",
            "retry_count": 3,
            "failed_at": "2017-10-17T16:49:09.647182+08:00"
        }
    }

### **POST /channels/{channel_name}/failed-tasks/{task_id}** 重试channel下失败的任务

#### 请求参数

| 参数 | 说明 |
|---|---|
| channel_name | 任务所在的channel |
| task_id | 任务ID |

#### 响应示例

    {
        "status_code": 200,
        "message": "ok",
        "data": null
    }

### **GET /delay-tasks** 查询所有延迟任务

### 响应示例

    {
        "status_code": 200,
        "message": "ok",
        "data": {
            "0a360b99-4b22-4269-bdec-9331deb8425c": {
                "task_id": "0a360b99-4b22-4269-bdec-9331deb8425c",
                "task_name": "git-command",
                "command": {
                    "name": "git",
                    "args": null
                },
                "channel": "default",
                "status": "",
                "execute_at": "2017-10-17T18:07:18.737976+08:00",
                "retry_count": 0,
                "failed_at": "0001-01-01T00:00:00Z"
            }
        }
    }

### **GET /delay-tasks/{task_id}** 查询某个延迟任务

#### 请求参数

| 参数 | 说明 |
|---|---|
| task_id | 任务ID |

#### 响应示例

    {
        "status_code": 200,
        "message": "ok",
        "data": {
            "task_id": "dcf8e5c8-3207-457d-b75c-23b4841734b4",
            "task_name": "git-command",
            "command": {
                "name": "git",
                "args": null
            },
            "channel": "default",
            "status": "",
            "execute_at": "2017-10-20T11:24:21.986208+08:00",
            "retry_count": 0,
            "failed_at": "0001-01-01T00:00:00Z"
        }
    }

### **DELETE /delay-tasks/{task_id}** 删除某个延迟任务

#### 请求参数

| 参数 | 说明 |
|---|---|
| task_id | 任务ID |

#### 响应示例

    {
        "status_code": 200,
        "message": "ok",
        "data": {
            "task_id": "182e1598-4915-4510-b2a9-3f9db8706dfd",
            "task_name": "git-command",
            "command": {
                "name": "git",
                "args": null
            },
            "channel": "default",
            "status": "",
            "execute_at": "2017-10-20T11:26:26.180545+08:00",
            "retry_count": 0,
            "failed_at": "0001-01-01T00:00:00Z"
        }
    }

## 注意事项

### 安全问题

**Coyotes** 会使用shell执行任务队列里的任何命令，包括系统相关的命令，如果不严格控制好输入，恶意用户可以非常轻松的拿到系统的root权限。

1. 一定不要直接开放给外网，**只在内部信任的网络使用**，比如绑定http地址到`192.168`开头的网卡或者`127.0.0.1`本地网络
2. 如果执行的命令包含来自用户的输入，一定要对输入内容进行严格检查，否则会引起命令注入，造成严重的安全问题，**不要相信任何来自用户的输入**
3. 不要使用**root权限**执行

## TODO

* [x] 解决退出信号会传递到子进程的问题
* [x] 队列channel持久化
* [ ] 支持单个channel的退出
* [ ] 排队中的任务取消
* [x] 实现队列调度器
* [ ] 使用可选后端记录任务执行结果、任务执行结果回调
* [ ] 实现RabbitMQ作为broker的支持
* [ ] 队列状态监控
* [ ] 集群支持（任务分发，新增队列）
* [ ] 支持脚本文件实时分发、执行
* [ ] 定时任务执行能力，发送一个任务到队列，设定一个执行时间，时间到了的时候再执行该任务（延迟执行）
* [ ] 增加配置文件
* [ ] 增加对Redis Sentinel和Redis Cluster的支持
* [ ] 解决可以执行任意系统命令带来的安全问题
* [ ] 加入Redis连接池的支持

