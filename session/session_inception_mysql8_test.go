// Copyright 2026 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package session_test

import (
	stdsql "database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hanchuanchuan/goInception/config"
	"github.com/hanchuanchuan/goInception/session"
	"golang.org/x/net/context"
)

// mysqlSourceOptions 从环境变量读取目标MySQL连接信息,默认与CI环境保持一致.
// 本地可用3307的MySQL8运行:
//
//	GOINCEPTION_TEST_MYSQL_PORT=3307 GOINCEPTION_TEST_MYSQL_USER=root \
//	GOINCEPTION_TEST_MYSQL_PASSWORD=xxx go test ./session \
//	-run TestMysql8ModifyTextColumnToVarchar -v
func mysqlSourceOptions(t *testing.T) session.SourceOptions {
	t.Helper()

	host := envOrDefault("GOINCEPTION_TEST_MYSQL_HOST", "127.0.0.1")
	port, err := strconv.Atoi(envOrDefault("GOINCEPTION_TEST_MYSQL_PORT", "3306"))
	if err != nil {
		t.Fatalf("非法的 GOINCEPTION_TEST_MYSQL_PORT: %v", err)
	}

	return session.SourceOptions{
		Host:         host,
		Port:         port,
		User:         envOrDefault("GOINCEPTION_TEST_MYSQL_USER", "test"),
		Password:     envOrDefault("GOINCEPTION_TEST_MYSQL_PASSWORD", "test"),
		RealRowCount: true,
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// TestMysql8ModifyTextColumnToVarchar 复现 TEXT 列修改为 VARCHAR(2048) 时
// checkModifyColumn 对旧类型做定长切片导致的越界 panic.
//
// 测试在真实MySQL中建表后,仅审核 ALTER 语句,使列类型元数据来自
// information_schema.COLUMN_TYPE (MySQL8中 text 即 "text", 长度只有4).
func TestMysql8ModifyTextColumnToVarchar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opt := mysqlSourceOptions(t)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		opt.User, opt.Password, opt.Host, opt.Port)

	db, err := stdsql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("连接MySQL失败: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("目标MySQL(%s:%d)不可用, 跳过: %v", opt.Host, opt.Port, err)
	}

	const (
		dbName    = "test_inc"
		tableName = "inception_test_text_to_varchar"
		alterSQL  = "ALTER TABLE " + dbName + "." + tableName +
			" MODIFY COLUMN recommend_info VARCHAR(2048) NULL COMMENT 'AI 输入快照：推荐语';"
	)

	mustExec(t, db, "CREATE DATABASE IF NOT EXISTS "+dbName+" DEFAULT CHARACTER SET utf8mb4")
	mustExec(t, db, "DROP TABLE IF EXISTS "+dbName+"."+tableName)
	defer db.Exec("DROP TABLE IF EXISTS " + dbName + "." + tableName)
	mustExec(t, db, "CREATE TABLE "+dbName+"."+tableName+" ("+
		"`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,"+
		"`recommend_info` TEXT NULL COMMENT 'AI 输入快照',"+
		"PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")

	inc := &config.GetGlobalConfig().Inc
	savedCheckColumnTypeChange := inc.CheckColumnTypeChange
	inc.CheckColumnTypeChange = true
	defer func() { inc.CheckColumnTypeChange = savedCheckColumnTypeChange }()

	se := session.NewInception()
	if err := se.LoadOptions(opt); err != nil {
		t.Fatalf("加载数据源失败: %v", err)
	}

	// 修复前此处会 panic: slice bounds out of range [:7] with capacity 4
	records, err := se.Audit(context.Background(), "use "+dbName+";\n"+alterSQL+"\n")
	if err != nil {
		t.Fatalf("审核执行失败: %v", err)
	}
	if len(records) == 0 {
		t.Fatal("审核未返回结果")
	}

	last := records[len(records)-1]
	wantLevel := session.GetErrorLevel(session.ER_CHANGE_COLUMN_TYPE)
	if last.ErrLevel != wantLevel {
		t.Fatalf("期望 %s 告警(级别%d), 实际级别%d, 信息: %s",
			session.ER_CHANGE_COLUMN_TYPE, wantLevel, last.ErrLevel, last.ErrorMessage)
	}
	if !strings.Contains(last.ErrorMessage, "varchar(2048)") ||
		!strings.Contains(last.ErrorMessage, "text") {
		t.Fatalf("告警信息未包含类型变更详情: %s", last.ErrorMessage)
	}
}

func mustExec(t *testing.T, db *stdsql.DB, query string) {
	t.Helper()
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("执行SQL失败: %v\nSQL: %s", err, query)
	}
}
