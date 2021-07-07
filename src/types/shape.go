package types

type ShapesMap map[string]ShapeInfo

type ShapeProp struct {
	Key   string
	Types []Type
}

type ShapeInfo struct {
	Name  string
	Props []ShapeProp
}
