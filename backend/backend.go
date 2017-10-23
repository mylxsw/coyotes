package backend

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/mylxsw/coyotes/brokers"
)

// Driver 后端存储接口
type Driver interface {
	// Insert 插入执行结果到数据库
	Insert(task brokers.Task, result Result) (ID string, err error)
	// ClearExpired 清理过期的历史记录
	ClearExpired(beforeTime time.Time) (cnt int64, err error)
}

// Result 任务执行后的结果
type Result struct {
	Stdout       string
	Stderr       string
	IsSuccessful bool
}

var drivers = make(map[string]Driver)
var defaultDriverName string
var driversMu sync.RWMutex

// Register 注册一个存储驱动
func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()

	if driver == nil {
		panic("storage: Register driver is nil")
	}

	if _, dup := drivers[name]; dup {
		panic("storage: Register called twice for driver " + name)
	}

	drivers[name] = driver

	if defaultDriverName == "" {
		defaultDriverName = name
	}
}

func unregisterAllDrivers() {
	driversMu.Lock()
	defer driversMu.Unlock()

	drivers = make(map[string]Driver)
}

// Drivers 返回所有已经注册了的驱动
func Drivers() []string {
	driversMu.RLock()
	defer driversMu.RUnlock()
	var list []string
	for name := range drivers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// Open 返回指定的驱动
func Open(driverName string) (Driver, error) {
	driversMu.RLock()
	driveri, ok := drivers[driverName]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("storage: unknown driver %q (forgotten import?)", driverName)
	}

	return driveri, nil
}

// Default 返回第一个注册的驱动
func Default() Driver {
	driver, _ := Open(defaultDriverName)
	return driver
}
