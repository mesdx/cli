package models

// type_usage_edges.go — fixture testing that CoreModel is found as a
// ref when it appears inside generic/container type expressions.
//
// Every function below uses CoreModel exclusively in type positions
// (parameters, return types, local variable declarations) so the fixture
// exercises Go's type-identifier capture (`(type_identifier) @ref.type`).

import (
	"fmt"
)

// TakesSlice receives a slice of CoreModel values.
func TakesSlice(models []CoreModel) int {
	return len(models)
}

// TakesPointerSlice receives a slice of CoreModel pointers.
func TakesPointerSlice(models []*CoreModel) int {
	return len(models)
}

// ReturnsPointer returns a pointer to a CoreModel.
func ReturnsPointer() *CoreModel {
	return NewCoreModel("edge", 0.5)
}

// TakesMap receives a map whose values are CoreModel.
func TakesMap(m map[string]CoreModel) int {
	return len(m)
}

// TakesMapOfPointers receives a map whose values are *CoreModel.
func TakesMapOfPointers(m map[string]*CoreModel) int {
	return len(m)
}

// TakesChan receives a channel of CoreModel.
func TakesChan(ch chan CoreModel) CoreModel {
	return <-ch
}

// LocalVarType declares a local variable of type CoreModel.
func LocalVarType() {
	var m CoreModel
	m.Title = "local"
	fmt.Println(m.Describe())
}

// LocalPointerType uses *CoreModel as a local variable type.
func LocalPointerType() {
	var p *CoreModel = NewCoreModel("ptr", 1.0)
	fmt.Println(p.IsValid())
}

// NestedSliceType uses [][]CoreModel.
func NestedSliceType(matrix [][]CoreModel) int {
	return len(matrix)
}

// FuncType uses CoreModel as both parameter and return type.
func FuncType(transform func(CoreModel) CoreModel) CoreModel {
	return transform(CoreModel{Title: "x", Score: 0})
}
