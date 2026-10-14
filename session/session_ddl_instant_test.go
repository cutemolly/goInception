// Copyright 2024 PingCAP, Inc.
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

package session

import (
	"testing"

	"github.com/hanchuanchuan/goInception/ast"
	"github.com/hanchuanchuan/goInception/parser"
	"github.com/hanchuanchuan/goInception/sessionctx/variable"
)

// Test_checkDDLInstantMatrix 覆盖 MySQL 8.0 ALGORITHM=INSTANT 判定矩阵.
// canInstant 为 true 时 goInception 会关闭 pt-osc/gh-ost, 走本地(in-place)执行.
func Test_checkDDLInstantMatrix(t *testing.T) {
	p := parser.New()
	sessionVars := variable.NewSessionVars()
	p.SetSQLMode(sessionVars.SQLMode)

	table := &TableInfo{Name: "t1",
		Fields: []FieldInfo{
			{Field: "c1", Extra: ""},
			{Field: "c2", Extra: "VIRTUAL"},
			{Field: "c3", Extra: "STORED"}}}

	tests := []struct {
		sql     string
		mysql57 bool
		mysql80 bool
		note    string
	}{
		// ---- 显式 ALGORITHM=INSTANT ----
		{"alter table t1 add column c9 int, algorithm=instant;", true, true, "显式 INSTANT"},
		{"alter table t1 modify column c1 bigint, algorithm=instant;", true, true, "显式 INSTANT 覆盖 MODIFY"},
		{"alter table t1 add index idx_c1(c1), algorithm=instant;", true, true, "显式 INSTANT 覆盖 ADD INDEX"},
		{"alter table t1 drop column c1, algorithm=instant;", true, true, "显式 INSTANT 覆盖 DROP COLUMN"},

		// ---- ADD COLUMN ----
		{"alter table t1 add column c9 int;", false, true, "普通 ADD COLUMN"},
		{"ALTER TABLE t1 ADD COLUMN c2 INT GENERATED ALWAYS AS (c1 + 1) VIRTUAL;", true, true, "VIRTUAL 生成列"},
		{"ALTER TABLE t1 ADD COLUMN c3 INT GENERATED ALWAYS AS (c1 + 1) STORED;", false, false, "STORED 生成列"},
		{"alter table t1 add column c4 int after c1;", false, false, "带 AFTER 定位(8.0.29前不支持)"},
		{"alter table t1 add column c4 int first;", false, false, "带 FIRST 定位(8.0.29前不支持)"},
		{"alter table t1 add column c9 int primary key;", false, false, "新增主键列"},
		{"alter table t1 add column c9 int unique;", false, false, "新增唯一列"},
		{"alter table t1 add column c9 int, add column c10 int;", false, true, "多列 ADD COLUMN"},

		// ---- DROP COLUMN ----
		{"alter table t1 drop column c1;", false, false, "普通 DROP COLUMN"},
		{"alter table t1 drop column c2;", false, true, "DROP VIRTUAL 列"},

		// ---- RENAME ----
		{"alter table t1 rename to t2;", false, true, "RENAME TO"},
		{"alter table t1 rename column c1 to c9;", false, false, "RENAME COLUMN(8.0 支持 INSTANT, 未识别)"},

		// ---- ALTER COLUMN (仅改默认值) ----
		{"alter table t1 alter column c1 set default 1;", false, true, "ALTER COLUMN SET DEFAULT"},

		// ---- ALTER INDEX (可见性切换, 8.0 只改元数据) ----
		{"alter table t1 alter index idx_c1 invisible;", false, true, "ALTER INDEX INVISIBLE"},
		{"alter table t1 alter index idx_c1 visible;", false, true, "ALTER INDEX VISIBLE"},
		{"alter table t1 add column c9 int, alter index idx_c1 invisible;", false, true, "ADD COLUMN + ALTER INDEX"},

		// ---- 不支持 INSTANT ----
		{"alter table t1 modify column c1 bigint;", false, false, "MODIFY COLUMN"},
		{"alter table t1 change column c1 c9 bigint;", false, false, "CHANGE COLUMN"},
		{"alter table t1 add index idx_c1(c1);", false, false, "ADD INDEX"},
		{"alter table t1 drop index idx_c1;", false, false, "DROP INDEX"},
		{"alter table t1 add primary key(c1);", false, false, "ADD PRIMARY KEY"},
		{"alter table t1 comment='x';", false, false, "表注释"},
		{"alter table t1 engine=innodb;", false, false, "改引擎"},
		{"alter table t1 force;", false, false, "FORCE"},

		// ---- 混合 ----
		{"alter table t1 add column c9 int, rename to t2;", false, true, "ADD COLUMN + RENAME"},
		{"alter table t1 add column c9 int, add index idx_c9(c9);", false, false, "ADD COLUMN + ADD INDEX"},
		{"alter table t1 add column c9 int, alter column c1 set default 1;", false, true, "ADD COLUMN + SET DEFAULT"},

		// ---- ALGORITHM / LOCK 组合 ----
		{"alter table t1 modify column c1 bigint, algorithm=inplace;", false, false, "显式 INPLACE + MODIFY"},
		{"alter table t1 modify column c1 bigint, algorithm=copy;", false, false, "显式 COPY + MODIFY"},
		{"alter table t1 add column c9 int, lock=none;", false, true, "LOCK=NONE + ADD COLUMN"},
	}

	for _, test := range tests {
		stmts, _, err := p.Parse(test.sql, "", "")
		if err != nil {
			t.Errorf("解析失败 %q: %v", test.sql, err)
			continue
		}
		if len(stmts) != 1 {
			t.Errorf("期望 1 条语句, 实际 %d: %q", len(stmts), test.sql)
			continue
		}
		node, ok := stmts[0].(*ast.AlterTableStmt)
		if !ok {
			t.Errorf("不是 AlterTableStmt: %q (%T)", test.sql, stmts[0])
			continue
		}

		// 解析出的 Algorithm 字段本身也要正确
		for _, spec := range node.Specs {
			if spec.Algorithm == ast.AlgorithmTypeInstant &&
				!hasInstantKeyword(test.sql) {
				t.Errorf("Algorithm 误判为 INSTANT: %q", test.sql)
			}
		}

		if got := checkDDLInstantMySQL57(node); got != test.mysql57 {
			t.Errorf("[MySQL5.7] %s: 期望 %v 实际 %v (%s)", test.sql, test.mysql57, got, test.note)
		}
		if got := checkDDLInstantMySQL80(node, table, 80000); got != test.mysql80 {
			t.Errorf("[MySQL8.0] %s: 期望 %v 实际 %v (%s)", test.sql, test.mysql80, got, test.note)
		}
	}
}

// Test_checkDDLInstantVersionBoundary 验证 8.0.29 版本边界(AFTER/FIRST 定位支持 INSTANT).
func Test_checkDDLInstantVersionBoundary(t *testing.T) {
	p := parser.New()
	sessionVars := variable.NewSessionVars()
	p.SetSQLMode(sessionVars.SQLMode)

	table := &TableInfo{Name: "t1", Fields: []FieldInfo{{Field: "c1"}}}

	stmts, _, err := p.Parse("alter table t1 add column c4 int after c1;", "", "")
	if err != nil {
		t.Fatal(err)
	}
	node := stmts[0].(*ast.AlterTableStmt)

	for _, tc := range []struct {
		version int
		want    bool
	}{
		{80000, false},
		{80028, false},
		{80029, true},
		{80035, true},
	} {
		if got := checkDDLInstantMySQL80(node, table, tc.version); got != tc.want {
			t.Errorf("dbVersion=%d: 期望 %v 实际 %v", tc.version, tc.want, got)
		}
	}
}

func hasInstantKeyword(sql string) bool {
	lower := ""
	for _, r := range sql {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		lower += string(r)
	}
	const kw = "algorithm=instant"
	for i := 0; i+len(kw) <= len(lower); i++ {
		if lower[i:i+len(kw)] == kw {
			return true
		}
	}
	return false
}
