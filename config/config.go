package config

// RedisConfig hold redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// HTTPConfig hold http configuration
type HTTPConfig struct {
	ListenAddr string
}

// Config hold all the configuration
type Config struct {
	Redis            RedisConfig
	HTTP             HTTPConfig
	Concurrent       int
	PidFile          string
	TaskMode         bool
	ColorfulTTY      bool
	DefaultChannel   string
	ChannelCacheSize int
}

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
	Task        chan Task   `json:"-"`
	Distinct    bool          `json:"distinct"`
	WorkerCount int           `json:"worker_count"`
	StopChan    chan struct{} `json:"-"`
}

// Runtime hold global runtime configuration
type Runtime struct {
	Config         Config
	Stoped         chan struct{}
	StopHTTPServer chan struct{}
	StopScheduler  chan struct{}
	Channels       map[string]*Channel
}

var runtime *Runtime

func InitRuntime(
	redisAddr string,
	redisPassword string,
	redisAddrDepressed string,
	redisPasswordDepressed string,
	pidFile string,
	concurrent int,
	redisDB int,
	httpAddr string,
	taskMode bool,
	colorfulTTY bool,
	defaultChannel string,
) *Runtime {

	if redisAddr == "127.0.0.1:6379" || redisAddr == "" {
		redisAddr = redisAddrDepressed
	}
	if redisPassword == "" {
		redisPassword = redisPasswordDepressed
	}

	runtime = &Runtime{
		Config: Config{
			PidFile:    pidFile,
			Concurrent: concurrent,
			Redis: RedisConfig{
				Addr:     redisAddr,
				Password: redisPassword,
				DB:       redisDB,
			},
			HTTP: HTTPConfig{
				ListenAddr: httpAddr,
			},
			TaskMode:         taskMode,
			ColorfulTTY:      colorfulTTY,
			DefaultChannel:   defaultChannel,
			ChannelCacheSize: 20,
		},
		Channels: make(map[string]*Channel),
	}

	// 用于向所有channel发送程序退出信号
	runtime.StopHTTPServer = make(chan struct{})
	runtime.StopScheduler = make(chan struct{})

	return runtime
}

// GetRuntime function return a runtime instance
func GetRuntime() *Runtime {
	return runtime
}
