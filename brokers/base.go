package brokers

import (
	"strconv"
	"strings"
	"time"
)

// PrepareTask is the task that prepared to join queue
type PrepareTask struct {
	ID        string      `json:"id"`
	Name      string      `json:"task"`
	Command   TaskCommand `json:"command"`
	Channel   string      `json:"chan"`
	Timestamp int64       `json:"ts"`
	ExecuteAt int64       `json:"execute_at"`
}

// Task represent a task object
type Task struct {
	ID           string      `json:"task_id"`
	TaskName     string      `json:"task_name"`
	Command      TaskCommand `json:"command"`
	Channel      string      `json:"channel"`
	Status       string      `json:"status"`
	ExecAt       time.Time   `json:"execute_at"`
	RetryCount   int         `json:"retry_count"`
	FailedAt     time.Time   `json:"failed_at"`
	WriteBackend bool        `json:"write_backend"`
}

// TaskCommand represent a task command object
type TaskCommand struct {
	Name string        `json:"name"`
	Args []interface{} `json:"args"`
}

// Channel is the command queue
type Channel struct {
	Name        string        `json:"name"`
	Task        chan Task     `json:"-"`
	Distinct    bool          `json:"distinct"`
	WorkerCount int           `json:"worker_count"`
	StopChan    chan struct{} `json:"-"`
	OutputChan  chan Output   `json:"-"`
}

// Output 任务执行输出
type Output struct {
	ProcessID string
	Type      string // 输出类型: stdout/stderr
	Task      Task
	Content   string
}

// Format format command struct to string
func (cmd TaskCommand) Format() string {
	return cmd.Name + " " + strings.Join(cmd.GetArgsString(), " ")
}

// GetArgsString 以字符串数组的形式返回参数集合
func (cmd TaskCommand) GetArgsString() []string {
	res := make([]string, len(cmd.Args))

	for i, s := range cmd.Args {
		switch s.(type) {
		case string:
			res[i] = s.(string)
		case float64:
			res[i] = strconv.FormatFloat(s.(float64), 'f', -1, 64)
		}
	}

	return res
}
