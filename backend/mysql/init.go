package mysql

import (
	"database/sql"
	"log"
)

// InitTableForMySQL 用于初始化MySQL表结构
func InitTableForMySQL(dataSourceName string) {
	tableCreateSQL := `
	CREATE TABLE IF NOT EXISTS histories (
		id int(10) unsigned NOT NULL AUTO_INCREMENT,
		task_name varchar(255) NOT NULL COMMENT '任务名', 
		command varchar(255) NOT NULL COMMENT '命令', 
		channel varchar(255) NOT NULL COMMENT '执行的队列', 
		status tinyint(1) DEFAULT '0' COMMENT '执行结果0-失败，1-成功', 
		retry_cnt int(11) unsigned DEFAULT '0', 
		stdout text COMMENT '标准输出', 
		stderr text COMMENT '标准错误输出',
		execute_at timestamp NULL DEFAULT NULL, 
		failed_at timestamp NULL DEFAULT NULL, 
		created_at timestamp NULL DEFAULT NULL,
		PRIMARY KEY (id)
	) DEFAULT CHARSET=utf8
	`

	initTable(tableCreateSQL, "mysql", dataSourceName)
}

func initTable(tableCreateSQL, name, dataSourceName string) {

	db, err := sql.Open(name, dataSourceName)
	if err != nil {
		log.Fatalf("Create db file failed: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(tableCreateSQL)
	if err != nil {
		log.Fatalf("Create table failed: %v", err)
	}
}
