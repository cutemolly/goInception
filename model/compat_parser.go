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

package model

// Types below are ported from the upstream tidb parser model package to keep
// the upgraded AST (parser/ast) compiling. They are referenced by AST nodes
// that goInception did not carry before, e.g. clustered index, foreign key
// refer options, stats column choice and resource group runaway options.

// PrimaryKeyType is the type of primary key.
// Available values are "clustered", "nonclustered", and ""(default).
type PrimaryKeyType int8

func (p PrimaryKeyType) String() string {
	switch p {
	case PrimaryKeyTypeClustered:
		return "CLUSTERED"
	case PrimaryKeyTypeNonClustered:
		return "NONCLUSTERED"
	default:
		return ""
	}
}

const (
	PrimaryKeyTypeDefault PrimaryKeyType = iota
	PrimaryKeyTypeClustered
	PrimaryKeyTypeNonClustered
)

// ReferOptionType is the type for refer options.
type ReferOptionType int

// Refer option types.
const (
	ReferOptionNoOption ReferOptionType = iota
	ReferOptionRestrict
	ReferOptionCascade
	ReferOptionSetNull
	ReferOptionNoAction
	ReferOptionSetDefault
)

// String implements fmt.Stringer interface.
func (r ReferOptionType) String() string {
	switch r {
	case ReferOptionRestrict:
		return "RESTRICT"
	case ReferOptionCascade:
		return "CASCADE"
	case ReferOptionSetNull:
		return "SET NULL"
	case ReferOptionNoAction:
		return "NO ACTION"
	case ReferOptionSetDefault:
		return "SET DEFAULT"
	}
	return ""
}

// ColumnChoice is the type of the column choice in the ANALYZE statement.
type ColumnChoice byte

const (
	DefaultChoice ColumnChoice = iota
	AllColumns
	PredicateColumns
	ColumnList
)

func (s ColumnChoice) String() string {
	switch s {
	case AllColumns:
		return "ALL"
	case PredicateColumns:
		return "PREDICATE"
	case ColumnList:
		return "LIST"
	default:
		return "DEFAULT"
	}
}

// Priority value of the resource group.
const (
	LowPriorityValue    = 1
	MediumPriorityValue = 8
	HighPriorityValue   = 16
)

// PriorityValueToName converts the priority value to corresponding name
func PriorityValueToName(value uint64) string {
	switch value {
	case LowPriorityValue:
		return "LOW"
	case MediumPriorityValue:
		return "MEDIUM"
	case HighPriorityValue:
		return "HIGH"
	default:
		return "MEDIUM"
	}
}

// RunawayActionType is the type of runaway action.
type RunawayActionType int32

const (
	RunawayActionNone RunawayActionType = iota
	RunawayActionDryRun
	RunawayActionCooldown
	RunawayActionKill
)

func (t RunawayActionType) String() string {
	switch t {
	case RunawayActionDryRun:
		return "DRYRUN"
	case RunawayActionCooldown:
		return "COOLDOWN"
	case RunawayActionKill:
		return "KILL"
	default:
		return "DRYRUN"
	}
}

// RunawayWatchType is the type of runaway watch.
type RunawayWatchType int32

const (
	WatchNone RunawayWatchType = iota
	WatchExact
	WatchSimilar
	WatchPlan
)

func (t RunawayWatchType) String() string {
	switch t {
	case WatchExact:
		return "EXACT"
	case WatchSimilar:
		return "SIMILAR"
	case WatchPlan:
		return "PLAN"
	default:
		return "NONE"
	}
}
