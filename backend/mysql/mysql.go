package mysql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	// 引入mysql驱动支持
	_ "github.com/go-sql-driver/mysql"
	"github.com/mylxsw/coyotes/backend"
	"github.com/mylxsw/coyotes/brokers"
)

// Storage 使用MySQL为存储引擎
type Storage struct {
	db             *sql.DB
	driverName     string
	dataSourceName string
}

// Register 注册当前驱动到Storage
func Register(driverName, dataSourceName string) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(fmt.Sprintf("%s: can not connect to db: %v", driverName, err))
	}

	backend.Register(driverName, &Storage{
		db:             db,
		driverName:     driverName,
		dataSourceName: dataSourceName,
	})
}

// Insert 插入执行结果到数据库
func (s *Storage) Insert(task brokers.Task, result backend.Result) (ID string, err error) {
	executeAt := "null"
	if !task.ExecAt.IsZero() {
		executeAt = "'" + task.ExecAt.Format("2006-01-02 15:04:05") + "'"
	}
	failedAt := "null"
	if !task.FailedAt.IsZero() {
		failedAt = "'" + task.FailedAt.Format("2006-01-02 15:04:05") + "'"
	}

	stmt, err := s.db.Prepare(fmt.Sprintf("INSERT INTO histories (task_name, command, channel, status, execute_at, retry_cnt, failed_at, stdout, stderr, created_at) VALUES(?, ?, ?, ?, %s, ?, %s, ?, ?, CURRENT_TIMESTAMP)", executeAt, failedAt))
	if err != nil {
		return "", fmt.Errorf("Prepare Error: %v", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		task.TaskName,
		task.Command.Name+" "+strings.Join(task.Command.Args, " "),
		task.Channel,
		result.IsSuccessful,
		task.RetryCount,
		result.Stdout,
		result.Stderr,
	)
	if err != nil {
		return "", fmt.Errorf("Insert Failed: %v", err)
	}

	lastInsertID, _ := res.LastInsertId()

	return strconv.Itoa(int(lastInsertID)), nil
}
