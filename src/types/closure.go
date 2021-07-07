package types

type ClosureMap map[string]ClosureInfo

type ClosureInfo struct {
	Name       string
	ReturnType []Type
	ParamTypes [][]Type
}
