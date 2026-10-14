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
	"golang.org/x/net/context"

	"github.com/hanchuanchuan/goInception/util/chunk"
)

// RecordSet is an abstract result set interface to help get data from Plan.
//
// The upstream parser removed RecordSet and Statement from the ast package,
// because they drag in util/chunk and context. goInception's executor, session
// and server layers are built around them, so they are kept here as a local
// extension.
type RecordSet interface {
	// Fields gets result fields.
	Fields() []*ResultField

	// Next reads records into chunk.
	Next(ctx context.Context, chk *chunk.Chunk) error

	// NewChunk creates a new chunk with initial capacity.
	NewChunk() *chunk.Chunk

	// Close closes the underlying iterator, call Next after Close will
	// restart the iteration.
	Close() error
}

// Statement is an interface for SQL execution.
// NOTE: all Statement implementations must be safe for
// concurrent using by multiple goroutines.
// If the Exec method requires any Execution domain local data,
// they must be held out of the implementing instance.
type Statement interface {
	// OriginText gets the origin SQL text.
	OriginText() string

	// Exec executes SQL and gets a Recordset.
	Exec(ctx context.Context) (RecordSet, error)

	// IsPrepared returns whether this statement is prepared statement.
	IsPrepared() bool

	// IsReadOnly returns if the statement is read only. For example: SelectStmt without lock.
	IsReadOnly() bool

	// RebuildPlan rebuilds the plan of the statement.
	RebuildPlan() (schemaVersion int64, err error)
}
