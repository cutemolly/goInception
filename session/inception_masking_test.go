// Copyright 2015 PingCAP, Inc.
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
	"fmt"
	"testing"

	"github.com/hanchuanchuan/goInception/config"
	"github.com/hanchuanchuan/goInception/util/testkit"
	. "github.com/pingcap/check"
)

var _ = Suite(&testSessionMaskingSuite{})

func TestMasking(t *testing.T) {
	TestingT(t)
}

type testSessionMaskingSuite struct {
	testCommon
}

func (s *testSessionMaskingSuite) SetUpSuite(c *C) {

	s.initSetUp(c)

	config.GetGlobalConfig().Inc.Lang = "en-US"
	config.GetGlobalConfig().Inc.EnableFingerprint = true
	config.GetGlobalConfig().Inc.SqlSafeUpdates = 0

	if s.tk == nil {
		s.tk = testkit.NewTestKitWithInit(c, s.store)
	}
}

func (s *testSessionMaskingSuite) TearDownSuite(c *C) {
	s.tearDownSuite(c)
}

func (s *testSessionMaskingSuite) makeSQL(sql string) *testkit.Result {
	a := `/*--user=test;--password=test;--host=127.0.0.1;--masking=1;--port=3306;--enable-ignore-warnings;*/
inception_magic_start;
use test_inc;
%s;
inception_magic_commit;`
	return s.tk.MustQueryInc(fmt.Sprintf(a, sql))
}

func (s *testSessionMaskingSuite) TestInsert(c *C) {

	res := s.makeSQL("insert into t1 values(1);")
	row := res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "2", Commentf("%v", row))
	c.Assert(row[4], Equals, "not support", Commentf("%v", row))

	res = s.makeSQL("insert into t1 values;")

	if len(res.Rows()) > 0 {
		row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
		c.Assert(row[2], Equals, "2", Commentf("%v", row))
		c.Assert(row[4], Equals, "line 1 column 21 near \"\" ", Commentf("%v", row))
	} else {
		c.Assert(len(res.Rows()), Greater, 0, Commentf("%v", res))
	}
}

func (s *testSessionMaskingSuite) TestSetOprBraces(c *C) {

	s.mustRunExec(c, `drop table if exists t1,t2;
	create table t1(id int primary key,c1 int);
	create table t2(id int primary key,c1 int);
	insert into t1 values(1,1),(2,2);`)

	// 括号包裹的操作数会被解析成 *ast.SetOprSelectList 嵌在 Selects 里,
	// 早期直接断言 *ast.SelectStmt 会 panic 并中断连接.
	cases := []string{
		`select id,c1 from t1 union (select id,c1 from t2)`,
		`(select id,c1 from t1) union (select id,c1 from t2)`,
		`(select id,c1 from t1 union select id,c1 from t2) union select 1,1`,
		`(select id,c1 from t1) except (select id,c1 from t2)`,
		`(select id,c1 from t1) intersect (select id,c1 from t2)`,
	}
	for _, sql := range cases {
		res := s.makeSQL(sql)
		row := res.Rows()[int(s.tk.Se.AffectedRows())-1]
		c.Assert(row[2], Equals, "0", Commentf("%s: %v", sql, row))
		// 两侧的表都要被脱敏字段覆盖, 漏掉括号操作数时 query_tree 会是空的
		c.Assert(row[3], Matches, `.*"table":"t1".*`, Commentf("%s: %v", sql, row))
		c.Assert(row[3], Matches, `.*"table":"t2".*`, Commentf("%s: %v", sql, row))
	}
}

func (s *testSessionMaskingSuite) TestQuery(c *C) {

	s.mustRunExec(c, `drop table if exists t1,t2;
	create table t1(id int primary key,c1 int);
	create table t2(id int primary key,c1 int,c2 int);
	insert into t1 values(1,1),(2,2);`)

	res := s.makeSQL(`select * from t1;`)
	row := res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, `[{"index":0,"field":"id","type":"int(11)","table":"t1","schema":"test_inc","alias":"id"},{"index":1,"field":"c1","type":"int(11)","table":"t1","schema":"test_inc","alias":"c1"}]`, Commentf("%v", row))

	res = s.makeSQL(`select a.* from t1 a;`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, `[{"index":0,"field":"id","type":"int(11)","table":"t1","schema":"test_inc","alias":"id"},{"index":1,"field":"c1","type":"int(11)","table":"t1","schema":"test_inc","alias":"c1"}]`, Commentf("%v", row))

	res = s.makeSQL(`select * from t1 union select id,c2 from t2;`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, `[{"index":0,"field":"id","type":"int(11)","table":"t1","schema":"test_inc","alias":"id"},{"index":1,"field":"c1","type":"int(11)","table":"t1","schema":"test_inc","alias":"c1"},{"index":2,"field":"id","type":"int(11)","table":"t2","schema":"test_inc","alias":"id"},{"index":3,"field":"c2","type":"int(11)","table":"t2","schema":"test_inc","alias":"c2"}]`, Commentf("%v", row))

	res = s.makeSQL(`select ifnull(c1,c2) from t1 inner join t2 on t1.id=t2.id`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, `[{"index":0,"field":"c1","type":"int(11)","table":"t1","schema":"test_inc","alias":"ifnull(c1,c2)"},{"index":0,"field":"c2","type":"int(11)","table":"t2","schema":"test_inc","alias":"ifnull(c1,c2)"}]`, Commentf("%v", row))

	res = s.makeSQL(`select a.c1_alias,a.c3_alias from (select *,c1 as c1_alias,concat(id,c2) as c3_alias from t2) a;`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, `[{"index":0,"field":"c1","type":"int(11)","table":"t2","schema":"test_inc","alias":"c1_alias"},{"index":1,"field":"id","type":"int(11)","table":"t2","schema":"test_inc","alias":"c3_alias"},{"index":1,"field":"c2","type":"int(11)","table":"t2","schema":"test_inc","alias":"c3_alias"}]`, Commentf("%v", row))

	res = s.makeSQL(`select a.*,concat(a.id,a.c1) as c4 from (select * from t2) a;`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, `[{"index":0,"field":"id","type":"int(11)","table":"t2","schema":"test_inc","alias":"id"},{"index":1,"field":"c1","type":"int(11)","table":"t2","schema":"test_inc","alias":"c1"},{"index":2,"field":"c2","type":"int(11)","table":"t2","schema":"test_inc","alias":"c2"},{"index":3,"field":"id","type":"int(11)","table":"t2","schema":"test_inc","alias":"c4"},{"index":3,"field":"c1","type":"int(11)","table":"t2","schema":"test_inc","alias":"c4"}]`, Commentf("%v", row))

	res = s.makeSQL(`select a1.*,a1.id,a2.* from t1 a1 inner join t2 a2 on a1.id=a2.id`)
	row = res.Rows()[int(s.tk.Se.AffectedRows())-1]
	c.Assert(row[2], Equals, "0", Commentf("%v", row))
	c.Assert(row[3], Equals, `[{"index":0,"field":"id","type":"int(11)","table":"t1","schema":"test_inc","alias":"id"},{"index":1,"field":"c1","type":"int(11)","table":"t1","schema":"test_inc","alias":"c1"},{"index":2,"field":"id","type":"int(11)","table":"t1","schema":"test_inc","alias":"id"},{"index":3,"field":"id","type":"int(11)","table":"t2","schema":"test_inc","alias":"id"},{"index":4,"field":"c1","type":"int(11)","table":"t2","schema":"test_inc","alias":"c1"},{"index":5,"field":"c2","type":"int(11)","table":"t2","schema":"test_inc","alias":"c2"}]`, Commentf("%v", row))

}
