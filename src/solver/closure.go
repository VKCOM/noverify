package solver

import (
	"github.com/VKCOM/noverify/src/meta"
)

// The model describes the types for the closure.
type ClosureArgsModel struct {
	Args []meta.TypesMap
}

func emptyModel() ClosureArgsModel {
	return ClosureArgsModel{}
}

// Class containing information about the function that called the closure.
type ClosureCallerInfo struct {
	FunctionName string          // caller function
	FunctionArgs []meta.TypesMap // types for each arg for call caller
}

// The model describes what types the arguments of the closure should have.
type ArgsModel struct {
	BaseTypeIndex  int // index of the element whose type will be the main one
	CountTypedArgs int // number of arguments in the callback that must have the type of the first element
	CountAllArgs   int // number of all arguments in the callback

	BaseTypeIndexShiftCount int // number of shifts for the base type. See the array_map function
}

// By function name and argument types, it returns a model that stores the argument types
// for the closure in the given function.
func (ci ClosureCallerInfo) Model() (ClosureArgsModel, bool) {
	switch ci.FunctionName {
	case `\usort`: // function usort(T[] $array, $callback) {}, $callback -> (T $a, T $b)
		return ci.model(ArgsModel{BaseTypeIndex: 0, CountTypedArgs: 2})
	case `\array_map`: // function array_map($callback, T[] $array) {}, $callback -> (T $a)
		if len(ci.FunctionArgs) > 2 { // array_map($callback, T[] $a1, T1[] $a2, [...]) {}, $callback -> (T $a, T1 $b, [...])
			count := len(ci.FunctionArgs) - 1
			return ci.model(ArgsModel{
				BaseTypeIndex:           1,
				CountTypedArgs:          1,
				BaseTypeIndexShiftCount: count,
			})
		}
		return ci.model(ArgsModel{BaseTypeIndex: 1, CountTypedArgs: 1})
	case `\array_walk`: // function array_walk(T[] $array, $callback) {}, $callback -> (T $value, mixed $key)
		return ci.model(ArgsModel{BaseTypeIndex: 0, CountTypedArgs: 1, CountAllArgs: 2})
	}

	return emptyModel(), false
}

func (ci ClosureCallerInfo) model(argsModel ArgsModel) (ClosureArgsModel, bool) {
	if argsModel.BaseTypeIndex >= len(ci.FunctionArgs) {
		return emptyModel(), false
	}

	if argsModel.CountAllArgs == 0 {
		argsModel.CountAllArgs = argsModel.CountTypedArgs
	}
	if argsModel.BaseTypeIndexShiftCount == 0 {
		argsModel.BaseTypeIndexShiftCount = 1
	}

	var args []meta.TypesMap

	for i := 0; i < argsModel.BaseTypeIndexShiftCount; i++ {
		tp, ok := ci.FunctionArgs[argsModel.BaseTypeIndex+i].ArrayBaseType()
		if !ok {
			return ClosureArgsModel{Args: args}, false
		}

		for i := 0; i < argsModel.CountTypedArgs; i++ {
			args = append(args, tp)
		}
	}

	if len(args) < argsModel.CountAllArgs {
		for i := 0; i < argsModel.CountAllArgs-len(args); i++ {
			args = append(args, meta.NewTypesMap("mixed"))
		}
	}

	return ClosureArgsModel{
		Args: args,
	}, true
}
