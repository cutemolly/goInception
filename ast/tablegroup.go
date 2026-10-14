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
	"strconv"

	"github.com/hanchuanchuan/goInception/format"
	"github.com/hanchuanchuan/goInception/model"
	"github.com/pingcap/errors"
)

// TABLEGROUP statements are an OceanBase compatibility feature carried by
// goInception. The upstream parser has no knowledge of them, so the AST nodes
// and the grammar rules that build them are kept as a local extension.

// TableGroupName represents a table group name.
type TableGroupName struct {
	node

	Name model.CIStr
}

// Restore implements Node interface.
func (n *TableGroupName) Restore(ctx *format.RestoreCtx) error {
	ctx.WriteName(n.Name.String())
	return nil
}

// Accept implements Node Accept interface.
func (n *TableGroupName) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*TableGroupName)
	return v.Leave(n)
}


type CreateTableGroupStmt struct {
	ddlNode

	IfNotExists bool
	TableGroup  *TableGroupName
	Options     []*TableGroupOption
	Partition   *TableGroupPartitionOption
}

// Restore implements Node interface.
func (n *CreateTableGroupStmt) Restore(ctx *format.RestoreCtx) error {
	ctx.WriteKeyWord("CREATE TABLEGROUP ")
	if n.IfNotExists {
		ctx.WriteKeyWord("IF NOT EXISTS ")
	}

	if err := n.TableGroup.Restore(ctx); err != nil {
		return errors.Annotate(err, "An error occurred while splicing CreateTableGroupStmt TableGroup")
	}

	for i, option := range n.Options {
		ctx.WritePlain(" ")
		if err := option.Restore(ctx); err != nil {
			return errors.Annotatef(err, "An error occurred while splicing CreateTableGroupStmt TableGroupOption: [%v]", i)
		}
	}

	if n.Partition != nil {
		ctx.WritePlain(" ")
		if err := n.Partition.Restore(ctx); err != nil {
			return errors.Annotate(err, "An error occurred while splicing CreateTableGroupStmt Partition")
		}
	}
	return nil
}

// Accept implements Node Accept interface.
func (n *CreateTableGroupStmt) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}

	n = newNode.(*CreateTableGroupStmt)
	node, ok := n.TableGroup.Accept(v)
	if !ok {
		return n, false
	}

	n.TableGroup = node.(*TableGroupName)
	if n.Partition != nil {
		node, ok := n.Partition.Accept(v)
		if !ok {
			return n, false
		}
		n.Partition = node.(*TableGroupPartitionOption)
	}

	return v.Leave(n)
}

// TableGroupOptionType is the type for TableGroupOption
type TableGroupOptionType int

// TableGroupOption types
const (
	TableGroupOptionNone TableGroupOptionType = iota
	TableGroupOptionLocality
	TableGroupOptionPrimaryZone
	TableGroupOptionTableGroupId
	TableGroupOptionBinding
	TableGroupOptionMaxUsedPartId
)

// TableGroupOption is used for parsing table group option from SQL.
type TableGroupOption struct {
	node

	Tp     TableGroupOptionType
	String string
	Num    uint64
	Bool   bool
}

// Restore implements Node interface.
func (n *TableGroupOption) Restore(ctx *format.RestoreCtx) error {
	switch n.Tp {
	case TableGroupOptionNone:
		return nil
	case TableGroupOptionLocality:
		ctx.WriteKeyWord("LOCALITY ")
		ctx.WriteString(n.String)
		return nil
	case TableGroupOptionPrimaryZone:
		ctx.WriteKeyWord("PRIMARY_ZONE ")
		ctx.WriteString(n.String)
		return nil
	case TableGroupOptionTableGroupId:
		ctx.WriteKeyWord("TABLEGROUP_ID ")
		ctx.WriteString(strconv.FormatUint(n.Num, 10))
		return nil
	case TableGroupOptionBinding:
		ctx.WriteKeyWord("BINDING ")
		ctx.WriteString(strconv.FormatBool(n.Bool))
		return nil
	case TableGroupOptionMaxUsedPartId:
		ctx.WriteKeyWord("MAX_USED_PART_ID ")
		ctx.WriteString(strconv.FormatUint(n.Num, 10))
		return nil
	default:
		return errors.Errorf("invalid TableGroupOptionType: %d", n.Tp)
	}
}

// Accept implements Node Accept interface.
func (n *TableGroupOption) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*TableGroupOption)
	return v.Leave(n)
}

// TableGroupPartitionOption specifies the table group partition options.
type TableGroupPartitionOption struct {
	node
	*PartitionMethod
	Sub         *PartitionMethod
	Definitions []*PartitionDefinition
}

// Validate checks if the partition is well-formed.
func (n *TableGroupPartitionOption) Validate() error {
	if len(n.Definitions) != 0 {
		if n.Num != 0 && n.Num != uint64(len(n.Definitions)) {
			return ErrPartitionWrongNoPart
		}
		subDefCount := len(n.Definitions[0].Sub)
		for _, subDef := range n.Definitions[1:] {
			if len(subDef.Sub) != subDefCount {
				return ErrPartitionWrongNoSubpart
			}
		}
	}

	switch n.Tp {
	case model.PartitionTypeHash, model.PartitionTypeKey:
		if n.Num == 0 {
			n.Num = 1
		}
	case model.PartitionTypeRange, model.PartitionTypeList:
		if len(n.Definitions) == 0 {
			return ErrPartitionsMustBeDefined.GenWithStackByArgs(n.Tp)
		}
	}
	return nil
}

// Restore implements Node interface.
func (n *TableGroupPartitionOption) Restore(ctx *format.RestoreCtx) error {
	ctx.WriteKeyWord("PARTITION BY " + n.PartitionMethod.Tp.String())

	if n.Sub != nil {
		ctx.WriteKeyWord(" SUBPARTITION BY " + n.Sub.Tp.String())
		if n.Sub.Num > 0 {
			ctx.WriteKeyWord(" SUBPARTITIONS ")
			ctx.WritePlainf("%d", n.Sub.Num)
		}
	}

	if n.Num > 0 {
		ctx.WriteKeyWord(" PARTITIONS ")
		ctx.WritePlainf("%d", n.Num)
	}

	if len(n.Definitions) > 0 {
		ctx.WritePlain(" (")
		for i, def := range n.Definitions {
			if i > 0 {
				ctx.WritePlain(",")
			}
			if err := def.Restore(ctx); err != nil {
				return errors.Annotatef(err, "An error occurred while restore PartitionOptions.Definitions[%d]", i)
			}
		}
		ctx.WritePlain(")")
	}
	return nil
}

// Accept implements Node Accept interface.
func (n *TableGroupPartitionOption) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}

	n = newNode.(*TableGroupPartitionOption)
	if !n.PartitionMethod.acceptInPlace(v) {
		return n, false
	}
	if n.Sub != nil && !n.Sub.acceptInPlace(v) {
		return n, false
	}
	for _, def := range n.Definitions {
		if !def.acceptInPlace(v) {
			return n, false
		}
	}
	return v.Leave(n)
}


type DropTableGroupStmt struct {
	ddlNode

	IfExists   bool
	TableGroup *TableGroupName
}

// Restore implements Node interface.
func (n *DropTableGroupStmt) Restore(ctx *format.RestoreCtx) error {
	ctx.WriteKeyWord("DROP TABLEGROUP ")
	if n.IfExists {
		ctx.WriteKeyWord("IF EXISTS ")
	}

	if err := n.TableGroup.Restore(ctx); err != nil {
		return errors.Annotate(err, "An error occurred while restore DropTableGroupStmt.TableGroup")
	}
	return nil
}

// Accept implements Node Accept interface.
func (n *DropTableGroupStmt) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*DropTableGroupStmt)
	node, ok := n.TableGroup.Accept(v)
	if !ok {
		return n, false
	}
	n.TableGroup = node.(*TableGroupName)
	return v.Leave(n)
}

type AlterTableGroupStmt struct {
	ddlNode

	TableGroup *TableGroupName
	Spec       *AlterTableGroupSpec
}

// Restore implements Node interface.
func (n *AlterTableGroupStmt) Restore(ctx *format.RestoreCtx) error {
	ctx.WriteKeyWord("ALTER TABLEGROUP ")
	if err := n.TableGroup.Restore(ctx); err != nil {
		return errors.Annotate(err, "An error occurred while restore AlterTableGroupStmt.TableGroup")
	}
	ctx.WritePlain(" ")
	if err := n.Spec.Restore(ctx); err != nil {
		return errors.Annotate(err, "An error occurred while restore AlterTableGroupStmt.Spec")
	}
	return nil
}

// Accept implements Node Accept interface.
func (n *AlterTableGroupStmt) Accept(v Visitor) (node Node, ok bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	if n.TableGroup != nil {
		node, ok := n.TableGroup.Accept(v)
		if !ok {
			return n, false
		}
		n.TableGroup = node.(*TableGroupName)
	}
	if n.Spec != nil {
		node, ok := n.Spec.Accept(v)
		if !ok {
			return n, false
		}
		n.Spec = node.(*AlterTableGroupSpec)
	}
	return v.Leave(n)
}

// AlterTableGroupSpec represents alter table group specification.
type AlterTableGroupSpec struct {
	node

	Tp             AlterTableGroupType
	Tables         []*TableName
	Options        []*TableGroupOption
	Partition      *TableGroupPartitionOption
	PartitionNames []model.CIStr
}

// Restore implements Node interface.
func (n *AlterTableGroupSpec) Restore(ctx *format.RestoreCtx) error {
	switch n.Tp {
	case AlterTableGroupTypeAddTable:
		ctx.WriteKeyWord("ADD TABLE ")
		for i, table := range n.Tables {
			if i != 0 {
				ctx.WritePlain(", ")
			}
			ctx.WriteString(table.Name.O)
		}
	case AlterTableGroupTypeSetOption:
		ctx.WriteKeyWord("SET ")
		for i, option := range n.Options {
			if i != 0 {
				ctx.WritePlain(", ")
			}
			if err := option.Restore(ctx); err != nil {
				return errors.Annotatef(err, "An error occurred while restore AlterTableGroupSpec.Options: [%v]", i)
			}
		}
	case AlterTableGroupTypeSetPartition:
		if err := n.Partition.Restore(ctx); err != nil {
			return errors.Annotate(err, "An error occurred while restore AlterTableGroupSpec.Partition")
		}
	case AlterTableGroupTypeDropPartition:
		ctx.WriteKeyWord("DROP PARTITION ")
		for i, name := range n.PartitionNames {
			if i != 0 {
				ctx.WritePlain(", ")
			}
			ctx.WritePlain(name.O)
		}
	case AlterTableGroupTypeDropSubPartition:
		ctx.WriteKeyWord("DROP SUBPARTITION ")
		for i, name := range n.PartitionNames {
			if i != 0 {
				ctx.WritePlain(", ")
			}
			ctx.WritePlain(name.O)
		}
	case AlterTableGroupTypeAddPartition:
		ctx.WriteKeyWord("ADD PARTITION (")
		for i, definition := range n.Partition.Definitions {
			if i != 0 {
				ctx.WritePlain(", ")
			}
			if err := definition.Restore(ctx); err != nil {
				return errors.Annotatef(err, "An error occurred while restore AlterTableGroupSpec.Partition.Definitions: [%v]", i)
			}
			ctx.WriteKeyWord(")")
		}
	case AlterTableGroupTypeReorganizePartition:
		ctx.WriteKeyWord("REORGANIZE PARTITION ")
		for i, name := range n.PartitionNames {
			if i != 0 {
				ctx.WritePlain(", ")
			}
			ctx.WritePlain(name.O)
		}
		ctx.WriteKeyWord(" INTO (")
		for i, definition := range n.Partition.Definitions {
			if i != 0 {
				ctx.WritePlain(", ")
			}
			if err := definition.Restore(ctx); err != nil {
				return errors.Annotatef(err, "An error occurred while restore AlterTableGroupSpec.Partition.Definitions: [%v]", i)
			}
			ctx.WriteKeyWord(")")
		}
	case AlterTableGroupTypeTruncatePartition:
		ctx.WriteKeyWord("TRUNCATE PARTITION ")
		for i, name := range n.PartitionNames {
			if i != 0 {
				ctx.WritePlain(", ")
			}
			ctx.WritePlain(name.O)
		}
	}
	return nil
}

// Accept implements Node Accept interface.
func (n *AlterTableGroupSpec) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*AlterTableGroupSpec)
	for i, table := range n.Tables {
		node, ok := table.Accept(v)
		if !ok {
			return n, false
		}
		n.Tables[i] = node.(*TableName)
	}
	for i, option := range n.Options {
		node, ok := option.Accept(v)
		if !ok {
			return n, false
		}
		n.Options[i] = node.(*TableGroupOption)
	}
	if n.Partition != nil && n.Partition.PartitionMethod != nil {
		node, ok := n.Partition.Accept(v)
		if !ok {
			return n, false
		}
		n.Partition = node.(*TableGroupPartitionOption)
	}
	return v.Leave(n)
}

// AlterTableGroupType is the type for AlterTableGroupStmt.
type AlterTableGroupType int

const (
	AlterTableGroupTypeAddTable = iota
	AlterTableGroupTypeSetOption
	AlterTableGroupTypeSetPartition
	AlterTableGroupTypeDropPartition
	AlterTableGroupTypeDropSubPartition
	AlterTableGroupTypeAddPartition
	AlterTableGroupTypeReorganizePartition
	AlterTableGroupTypeTruncatePartition
)
