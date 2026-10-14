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

package ast

import (
	. "github.com/hanchuanchuan/goInception/format"
)

var (
	_ StmtNode = &InceptionStartStmt{}
	_ StmtNode = &InceptionCommitStmt{}
	_ StmtNode = &InceptionSetStmt{}

	_ StmtNode = &ShowOscStmt{}
)

// OscOptionType is the operation type of a show osc statement.
type OscOptionType int

const (
	// OscOptionNone shows osc process information.
	OscOptionNone OscOptionType = iota
	// OscOptionKill kills the osc process.
	OscOptionKill
	// OscOptionPause pauses the osc process (gh-ost only).
	OscOptionPause
	// OscOptionResume resumes the osc process (gh-ost only).
	OscOptionResume
)

const (
	// ShowLevels is the show statement type of `show levels`.
	ShowLevels ShowStmtType = 1001
)

// ShowOscStmt is the AST node of pt-osc and gh-ost related statements.
type ShowOscStmt struct {
	stmtNode

	Sqlsha1 string

	Tp OscOptionType
}

// Restore implements Node interface.
func (n *ShowOscStmt) Restore(ctx *RestoreCtx) error {
	return nil
}

// Accept implements Node Accept interface.
func (n *ShowOscStmt) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*ShowOscStmt)
	return v.Leave(n)
}

// InceptionSetStmt is the statement to set inception variables.
type InceptionSetStmt struct {
	stmtNode
	// Variables is the list of variable assignment.
	Variables []*VariableAssignment
}

// Restore implements Node interface.
func (n *InceptionSetStmt) Restore(ctx *RestoreCtx) error {
	return nil
}

// Accept implements Node Accept interface.
func (n *InceptionSetStmt) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*InceptionSetStmt)
	for i, val := range n.Variables {
		node, ok := val.Accept(v)
		if !ok {
			return n, false
		}
		n.Variables[i] = node.(*VariableAssignment)
	}
	return v.Leave(n)
}

// InceptionStartStmt is a statement to start an inception session.
type InceptionStartStmt struct {
	stmtNode
}

// Restore implements Node interface.
func (n *InceptionStartStmt) Restore(ctx *RestoreCtx) error {
	return nil
}

// Accept implements Node Accept interface.
func (n *InceptionStartStmt) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*InceptionStartStmt)
	return v.Leave(n)
}

// InceptionCommitStmt is a statement to commit the current inception session.
type InceptionCommitStmt struct {
	stmtNode
}

// Restore implements Node interface.
func (n *InceptionCommitStmt) Restore(ctx *RestoreCtx) error {
	return nil
}

// Accept implements Node Accept interface.
func (n *InceptionCommitStmt) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*InceptionCommitStmt)
	return v.Leave(n)
}
