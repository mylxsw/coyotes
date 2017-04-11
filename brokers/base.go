package brokers

// PrepareTask is the task that prepared to join queue
type PrepareTask struct {
	Name      string `json:"task"`
	Channel   string `json:"chan"`
	Timestamp int64  `json:"ts"`
}

// Task represent a task object
type Task struct {
	ID       string `json:"task_id"`
	TaskName string `json:"task_name"`
	Channel  string `json:"channel"`
	Status   string `json:"status"`
}

// Channel is the command queue
type Channel struct {
	Name        string        `json:"name"`
	Task        chan Task     `json:"-"`
	Distinct    bool          `json:"distinct"`
	WorkerCount int           `json:"worker_count"`
	StopChan    chan struct{} `json:"-"`
}