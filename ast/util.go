// Copyright 2018 PingCAP, Inc.
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

package ast

import "math"

// UnspecifiedSize is unspecified size.
const (
	UnspecifiedSize = math.MaxUint64
)

// SetOprSelects 按出现顺序展平 set 操作(UNION/EXCEPT/INTERSECT)的操作数.
//
// 上游解析器把 SetOprSelectList.Selects 声明为 []Node: 只有括号包裹的操作数
// 才会被包装成 *SetOprSelectList 再嵌一层, 例如
//
//	SELECT 1 UNION SELECT 2                  -> [SelectStmt, SelectStmt]
//	(SELECT 1) UNION (SELECT 2)              -> [SetOprSelectList, SetOprSelectList]
//	SELECT 1 UNION (SELECT 2 UNION SELECT 3) -> [SelectStmt, SetOprSelectList]
//
// 旧解析器是强类型 []*SelectStmt, 括号只体现在 SelectStmt.IsInBraces 上.
// 因此遍历 Selects 时直接断言 *SelectStmt 会静默丢掉整个括号操作数(审核漏检),
// 无保护的断言则会 panic, 一律改用本函数.
func SetOprSelects(list *SetOprSelectList) []*SelectStmt {
	if list == nil {
		return nil
	}
	selects := make([]*SelectStmt, 0, len(list.Selects))
	for _, sel := range list.Selects {
		switch x := sel.(type) {
		case *SelectStmt:
			selects = append(selects, x)
		case *SetOprSelectList:
			selects = append(selects, SetOprSelects(x)...)
		}
	}
	return selects
}

// IsReadOnly checks whether the input ast is readOnly.
func IsReadOnly(node Node) bool {
	switch st := node.(type) {
	case *SelectStmt:
		if st.LockInfo != nil {
			switch st.LockInfo.LockType {
			case SelectLockForUpdate, SelectLockForUpdateNoWait, SelectLockForUpdateWaitN:
				return false
			}
		}

		checker := readOnlyChecker{
			readOnly: true,
		}

		node.Accept(&checker)
		return checker.readOnly
	case *ExplainStmt:
		return !st.Analyze || IsReadOnly(st.Stmt)
	case *DoStmt, *ShowStmt:
		return true
	case *SetOprStmt:
		for _, sel := range node.(*SetOprStmt).SelectList.Selects {
			if !IsReadOnly(sel) {
				return false
			}
		}
		return true
	case *SetOprSelectList:
		for _, sel := range node.(*SetOprSelectList).Selects {
			if !IsReadOnly(sel) {
				return false
			}
		}
		return true
	case *AdminStmt:
		switch node.(*AdminStmt).Tp {
		case AdminShowTelemetry, AdminShowDDL, AdminShowDDLJobs, AdminShowSlow,
			AdminCaptureBindings, AdminShowNextRowID, AdminShowDDLJobQueries,
			AdminShowDDLJobQueriesWithRange:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// readOnlyChecker checks whether a query's ast is readonly, if it satisfied
// 1. selectstmt;
// 2. need not to set var;
// it is readonly statement.
type readOnlyChecker struct {
	readOnly bool
}

// Enter implements Visitor interface.
func (checker *readOnlyChecker) Enter(in Node) (out Node, skipChildren bool) {
	if node, ok := in.(*VariableExpr); ok {
		// like func rewriteVariable(), this stands for SetVar.
		if !node.IsSystem && node.Value != nil {
			checker.readOnly = false
			return in, true
		}
	}
	return in, false
}

// Leave implements Visitor interface.
func (checker *readOnlyChecker) Leave(in Node) (out Node, ok bool) {
	return in, checker.readOnly
}
