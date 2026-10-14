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

package types

// FieldType accessors. The upstream parser AST and grammar access FieldType
// through getters/setters, while this fork exposes the fields directly. These
// wrappers keep the upgraded parser compiling without changing the field
// layout used by the rest of goInception.

// GetType returns the type of the FieldType.
func (ft *FieldType) GetType() byte {
	return ft.Tp
}

// SetType sets the type of the FieldType.
func (ft *FieldType) SetType(tp byte) {
	ft.Tp = tp
	ft.Array = false
}

// GetFlag returns the flag of the FieldType.
func (ft *FieldType) GetFlag() uint {
	return ft.Flag
}

// SetFlag sets the flag of the FieldType.
func (ft *FieldType) SetFlag(flag uint) {
	ft.Flag = flag
}

// AddFlag adds a flag to the FieldType.
func (ft *FieldType) AddFlag(flag uint) {
	ft.Flag |= flag
}

// AndFlag and the flag of the FieldType.
func (ft *FieldType) AndFlag(flag uint) {
	ft.Flag &= flag
}

// ToggleFlag toggle the flag of the FieldType.
func (ft *FieldType) ToggleFlag(flag uint) {
	ft.Flag ^= flag
}

// DelFlag deletes a flag from the FieldType.
func (ft *FieldType) DelFlag(flag uint) {
	ft.Flag &= ^flag
}

// GetFlen returns the flen of the FieldType.
func (ft *FieldType) GetFlen() int {
	return ft.Flen
}

// SetFlen sets the flen of the FieldType.
func (ft *FieldType) SetFlen(flen int) {
	ft.Flen = flen
}

// GetDecimal returns the decimal of the FieldType.
func (ft *FieldType) GetDecimal() int {
	return ft.Decimal
}

// SetDecimal sets the decimal of the FieldType.
func (ft *FieldType) SetDecimal(decimal int) {
	ft.Decimal = decimal
}

// GetCharset returns the charset of the FieldType.
func (ft *FieldType) GetCharset() string {
	return ft.Charset
}

// SetCharset sets the charset of the FieldType.
func (ft *FieldType) SetCharset(charset string) {
	ft.Charset = charset
}

// GetCollate returns the collate of the FieldType.
func (ft *FieldType) GetCollate() string {
	return ft.Collate
}

// SetCollate sets the collate of the FieldType.
func (ft *FieldType) SetCollate(collate string) {
	ft.Collate = collate
}

// GetElems returns the elements of the FieldType.
func (ft *FieldType) GetElems() []string {
	return ft.Elems
}

// SetElems sets the elements of the FieldType.
func (ft *FieldType) SetElems(elems []string) {
	ft.Elems = elems
}

// SetElem sets the element of the FieldType.
func (ft *FieldType) SetElem(idx int, element string) {
	ft.Elems[idx] = element
}

// GetElem returns the element of the FieldType.
func (ft *FieldType) GetElem(idx int) string {
	return ft.Elems[idx]
}

// SetArray sets the array field of the FieldType.
func (ft *FieldType) SetArray(array bool) {
	ft.Array = array
}

// IsArray return true if the field type is array.
func (ft *FieldType) IsArray() bool {
	return ft.Array
}

// ArrayType returns the element type of the array.
func (ft *FieldType) ArrayType() *FieldType {
	if !ft.Array {
		return ft
	}
	nt := *ft
	nt.Array = false
	return &nt
}

// SetElemWithIsBinaryLit sets the element of the FieldType and records whether
// it comes from a binary literal.
func (ft *FieldType) SetElemWithIsBinaryLit(idx int, element string, isBinaryLit bool) {
	ft.Elems[idx] = element
	if isBinaryLit {
		// Create the binary literal flags lazily.
		if ft.ElemsIsBinaryLit == nil {
			ft.ElemsIsBinaryLit = make([]bool, len(ft.Elems))
		}
		ft.ElemsIsBinaryLit[idx] = true
	}
}

// GetElemIsBinaryLit returns the binary literal flag of the element at index idx.
func (ft *FieldType) GetElemIsBinaryLit(idx int) bool {
	if len(ft.ElemsIsBinaryLit) == 0 {
		return false
	}
	return ft.ElemsIsBinaryLit[idx]
}

// CleanElemIsBinaryLit cleans the binary literal flag of all the elements.
func (ft *FieldType) CleanElemIsBinaryLit() {
	if ft != nil && ft.ElemsIsBinaryLit != nil {
		ft.ElemsIsBinaryLit = nil
	}
}
