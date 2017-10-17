package brokers

import "time"

// PrepareTask is the task that prepared to join queue
type PrepareTask struct {
	Name      string      `json:"task"`
	Command   TaskCommand `json:"command"`
	Channel   string      `json:"chan"`
	Timestamp int64       `json:"ts"`
}

// Task represent a task object
type Task struct {
	ID         string      `json:"task_id"`
	TaskName   string      `json:"task_name"`
	Command    TaskCommand `json:"command"`
	Channel    string      `json:"channel"`
	Status     string      `json:"status"`
	RetryCount int         `json:"retry_count"`
	FailedAt   time.Time   `json:"failed_at"`
}

// TaskCommand represent a task command object
type TaskCommand struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
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
	Task      Task
	Content   string
}
