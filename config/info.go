package config

import "sync"

var lock sync.Mutex

// IncrSuccTaskCount 增加执行成功的任务数量
func IncrSuccTaskCount() {
	lock.Lock()
	defer lock.Unlock()

	runtime.Info.SuccTaskCount ++
	runtime.Info.DealTaskCount ++
}

// IncrFailTaskCount 增加执行失败的任务数量
func IncrFailTaskCount() {
	lock.Lock()
	defer lock.Unlock()

	runtime.Info.FailTaskCount ++
	runtime.Info.DealTaskCount ++
}
